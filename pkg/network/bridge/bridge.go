/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package bridge

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"slices"
	"sort"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"peta.io/peta/pkg/network"
	"peta.io/peta/pkg/network/ip"
	"peta.io/peta/pkg/network/ipam"
	"peta.io/peta/pkg/network/netlinksafe"
	"peta.io/peta/pkg/utils/sysctl"
)

const defaultBrName = "br0"

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

type gwInfo struct {
	gws               []net.IPNet
	family            int
	defaultRouteFound bool
}

type cniBridgeIf struct {
	Name        string
	ifIndex     int
	peerIndex   int
	masterIndex int
	found       bool
}

type VlanTrunk struct {
	MinID *int `json:"minID,omitempty" yaml:"minID,omitempty"`
	MaxID *int `json:"maxID,omitempty" yaml:"maxID,omitempty"`
	ID    *int `json:"id,omitempty" yaml:"id,omitempty"`
}

type BridgeArgs struct {
	Mac string `json:"mac,omitempty" yaml:"mac,omitempty"`
}

type BrConf struct {
	network.Conf
	BrName string `json:"bridge" yaml:"bridge"`

	// To assign an IP address to the bridge device and enable IP forwarding
	// (e.g. 10.244.0.1 for the subnet 10.244.0.0/24)
	IsGW bool `json:"isGateway" yaml:"isGateway"`

	// Change the default route of the host machine to point to the bridge IP
	IsDefaultGW bool `json:"isDefaultGateway" yaml:"isDefaultGateway"`

	ForceAddress bool `json:"forceAddress" yaml:"forceAddress"`

	IPMasq        bool    `json:"ipMasq" yaml:"ipMasq"`
	IPMasqBackend *string `json:"ipMasqBackend,omitempty" yaml:"IPMasqBackend"` // "iptables" or "nftables"
	MTU           int     `json:"mtu" yaml:"mtu"`
	HairpinMode   bool    `json:"hairpinMode" yaml:"hairpinMode"`
	PromiscMode   bool    `json:"promiscMode" yaml:"promiscMode"`

	// VLAN Mode
	Vlan                int          `json:"vlan" yaml:"vlan"`
	VlanTrunk           []*VlanTrunk `json:"vlanTrunk,omitempty" yaml:"vlanTrunk,omitempty"`
	PreserveDefaultVlan bool         `json:"preserveDefaultVlan" yaml:"preserveDefaultVlan"`

	MacSpoofChk               bool `json:"macspoofchk,omitempty" yaml:"macspoofchk,omitempty"`
	EnableDad                 bool `json:"enabledad,omitempty" yaml:"enabledad,omitempty"` //Enable IPv6 Duplicate Address Detection (DAD)
	DisableContainerInterface bool `json:"disableContainerInterface,omitempty" yaml:"disableContainerInterface,omitempty"`
	PortIsolation             bool `json:"portIsolation,omitempty" yaml:"portIsolation,omitempty"`

	Args struct {
		Cni BridgeArgs `json:"cni,omitempty" yaml:"cni,omitempty"`
	} `json:"args,omitempty" yaml:"args,omitempty"`
	RuntimeConfig struct {
		Mac string `json:"mac,omitempty" yaml:"mac,omitempty"`
	} `json:"runtimeConfig,omitempty" yaml:"runtimeConfig,omitempty"`

	mac   string
	vlans []int
}

func SetupBridge(h netlinksafe.Handle, c *BrConf) (*netlink.Bridge, error) {
	var err error
	c, _, err = loadBrConf(c)
	if err != nil {
		return nil, err
	}
	// 1. Check config
	isLayer3 := c.IPAM.Type != ""

	if isLayer3 && c.DisableContainerInterface {
		return nil, fmt.Errorf("cannot use IPAM when DisableContainerInterface flag is set")
	}

	if c.IsDefaultGW {
		c.IsGW = true
	}

	if c.HairpinMode && c.PromiscMode {
		return nil, fmt.Errorf("cannot set hairpin mode and promiscuous mode at the same time")
	}

	// 2. Enable IP forward
	if isLayer3 {

		err = ip.EnableIP4Forward()
		if err != nil {
			return nil, err
		}
		err = ip.EnableIP6Forward()
		if err != nil {
			return nil, err
		}
	}

	// 3. Create bridge
	br, err := setupBridge(h, c)
	if err != nil {
		return nil, err
	}

	// 4. Set IPAM
	if isLayer3 {
		subnets := c.IPAM.IPAMSpec.PodCIDR
		var gws []*net.IPNet
		list, err := h.AddrStrList(br, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		for _, subnet := range subnets {
			gw, err := calcGateways(subnet)
			if err != nil {
				return nil, err
			}
			// config br ip
			addr := &netlink.Addr{
				IPNet: gw,
			}

			if !slices.Contains(list, gw.String()) {
				gws = append(gws, addr.IPNet)
				if err := h.AddrAdd(br, addr); err != nil {
					return nil, fmt.Errorf("error adding IP address(%s) to bridge: %v", addr.IP.String(), err)
				}
			}
		}

		// IP Masquerade
		if c.IPMasq {
			err := ip.SetupIPMasqForNetworks(c.IPMasqBackend, gws, c.Name, "", "")
			if err != nil {
				return nil, err
			}
		}
	}

	return br, nil
}

func setupBridge(h netlinksafe.Handle, c *BrConf) (*netlink.Bridge, error) {
	vlanFiltering := c.Vlan != 0 || c.VlanTrunk != nil
	br, err := ensureBridge(h, c.BrName, c.MTU, c.PromiscMode, vlanFiltering)
	if err != nil {
		return nil, fmt.Errorf("failed to create bridge %q: %v", c.BrName, err)
	}
	return br, err
}

func ensureBridge(h netlinksafe.Handle, brName string, mtu int, promiscMode, vlanFiltering bool) (*netlink.Bridge, error) {
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = brName
	linkAttrs.MTU = mtu
	br := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}
	if vlanFiltering {
		br.VlanFiltering = &vlanFiltering
	}

	err := h.LinkAdd(br)
	if err != nil {
		if errors.Is(err, unix.EEXIST) {
			// Modify the exist bridge
			err := h.LinkModify(br)
			if err != nil {
				return nil, fmt.Errorf("could not modify %q: %v", brName, err)
			}
		} else {
			return nil, fmt.Errorf("could not add %q: %v", brName, err)
		}
	}

	if promiscMode {
		if err := h.SetPromiscOn(br); err != nil {
			return nil, fmt.Errorf("could not set promiscuous mode on %q: %v", brName, err)
		}
	}

	// Re-fetch link to read all attrs and if it already existed,
	// ensure it's really a bridge with similar configuration
	br, err = bridgeByName(h, brName)
	if err != nil {
		return nil, err
	}

	// we want to own the ipv6 routes for this interface
	_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv6/conf/%s/accept_ra", brName))

	// The bridge must be associated with at least one "active" physical/virtual device
	if err := h.LinkSetUp(br); err != nil {
		return nil, err
	}

	return br, nil
}

func ensureVlanInterface(h netlinksafe.Handle, br *netlink.Bridge, vlanID int, preserveDefaultVlan bool) (netlink.Link, error) {
	name := fmt.Sprintf("%s.%d", br.Name, vlanID)

	brGatewayVeth, err := h.LinkByName(name)
	if err != nil {
		if !errors.As(err, &netlink.LinkNotFoundError{}) || err.Error() != "link not found" {
			return nil, fmt.Errorf("could not lookup vlan-interface %q: %v", name, err)
		}

		vlan, err := setupVlan(h, br, name, br.MTU, false, vlanID, nil, preserveDefaultVlan, "", false)
		if err != nil {
			return nil, fmt.Errorf("could not setup vlan-interface %q: %v", name, err)
		}

		brGatewayVeth, err = h.LinkByName(vlan.Name)
		if err != nil {
			return nil, fmt.Errorf("could not lookup vlan-interface %q: %v", vlan.Name, err)
		}

		err = h.LinkSetUp(brGatewayVeth)
		if err != nil {
			return nil, fmt.Errorf("could not set vlan-interface %q: %v", vlan.Name, err)
		}
	}
	return brGatewayVeth, nil
}

func setupVlan(
	h netlinksafe.Handle,
	br *netlink.Bridge,
	ifName string,
	mtu int,
	harpinMode bool,
	vlanID int,
	vlans []int,
	preserveDefaultVlan bool,
	mac string,
	portIsolation bool,
) (*netlink.Vlan, error) {
	// ifName: Vlan name
	vlan := &netlink.Vlan{}
	return vlan, nil
}

func bridgeByName(h netlinksafe.Handle, name string) (*netlink.Bridge, error) {
	l, err := h.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not lookup %q: %w", name, err)
	}
	br, ok := l.(*netlink.Bridge)
	if !ok {
		return nil, fmt.Errorf("%q already exists but is not a bridge", name)
	}
	return br, nil
}

func calcGateways(cidr string) (*net.IPNet, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("error parsing subnet %q: %v", cidr, err)
	}

	gw, err := ipam.GetIndexedIP(ipNet, 1)
	if err != nil {
		return nil, fmt.Errorf("error getting gateway for subnet %q: %v", cidr, err)
	}
	return &net.IPNet{
		IP:   gw,
		Mask: ipNet.Mask,
	}, nil
}

func addBrAddr(h netlinksafe.Handle, br *netlink.Bridge, addr net.Addr) error {
	// Check addr exists
	//	addr := &netlink.Addr{
	//		IPNet: gw,
	//	}
	//
	//	if err := h.AddrAdd(br, addr); err != nil {
	//		return nil, fmt.Errorf("error adding IP address(%s) to bridge: %v", addr.IP.String(), err)
	//	}
	//}
	//addresses, err := h.AddrList(br, netlink.FAMILY_V4)
	//if err != nil {
	//return nil, err
	//}
	//for _, addr := range addresses {
	//fmt.Printf(" - %s\n", addr.String())
	//}
	addresses, err := h.AddrList(br, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("error listing addresses for bridge %q: %v", br.Name, err)
	}
	// addresses = [1, 2, 3, 4]
	// bSlice = [3, 4, 5]
	for _, a := range addresses {
		if a.String() != addr.String() {

		}
	}
	return nil
}

func loadBrConf(c *BrConf) (*BrConf, string, error) {

	if c.BrName == "" {
		c.BrName = defaultBrName
	}

	if c.Vlan < 0 || c.Vlan > 4094 {
		return nil, "", fmt.Errorf("invalid VLAN ID %d (must be between 0 and 4094)", c.Vlan)
	}

	var err error
	c.vlans, err = collectVlanTrunk(c.VlanTrunk)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse vlan trunks: %v", err)
	}

	if mac := c.RuntimeConfig.Mac; mac != "" {
		c.mac = mac
	}

	return c, c.CNIVersion, nil
}

func collectVlanTrunk(vlanTrunk []*VlanTrunk) ([]int, error) {
	if vlanTrunk == nil {
		return nil, nil
	}

	vlanMap := make(map[int]struct{})
	for _, item := range vlanTrunk {
		var minID int
		var maxID int
		var ID int

		switch {
		case item.MinID != nil && item.MaxID != nil:
			minID = *item.MinID
			if minID <= 0 || minID > 4094 {
				return nil, errors.New("incorrect trunk minID parameter")
			}
			maxID = *item.MaxID
			if maxID <= 0 || maxID > 4094 {
				return nil, errors.New("incorrect trunk maxID parameter")
			}
			if maxID < minID {
				return nil, errors.New("minID is greater than maxID in trunk parameter")
			}
			for v := minID; v <= maxID; v++ {
				vlanMap[v] = struct{}{}
			}
		case item.MinID == nil && item.MaxID != nil:
			return nil, errors.New("minID and maxID should be configured simultaneously, minID is missing")
		case item.MinID != nil && item.MaxID == nil:
			return nil, errors.New("minID and maxID should be configured simultaneously, maxID is missing")
		}

		// single vid
		if item.ID != nil {
			ID = *item.ID
			if ID <= 0 || ID > 4094 {
				return nil, errors.New("incorrect trunk id parameter")
			}
			vlanMap[ID] = struct{}{}
		}
	}

	if len(vlanMap) == 0 {
		return nil, nil
	}
	vlans := make([]int, 0, len(vlanMap))
	for k := range vlanMap {
		vlans = append(vlans, k)
	}
	sort.Slice(vlans, func(i int, j int) bool { return vlans[i] < vlans[j] })
	return vlans, nil
}
