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
	"fmt"
	"slices"
	"testing"

	"github.com/vishvananda/netlink"
	"peta.io/peta/pkg/network"
	"peta.io/peta/pkg/network/netlinksafe"
)

func TestBridge(t *testing.T) {
	h, err := netlinksafe.NewHandle()
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	IPMasqBackend := "iptables"
	c := &BrConf{
		Conf: network.Conf{
			CNIVersion:   "1.0.0",
			Name:         "peta-cni",
			Type:         "bridge",
			Capabilities: make(map[string]bool),
			IPAM: network.IPAM{
				Type: "host-scope",
				IPAMSpec: network.IPAMSpec{
					PodCIDR: []string{
						"10.20.1.0/24",
						"10.20.2.0/24",
						"10.20.3.0/24",
					},
					MinAllocate:       0,
					MaxAllocate:       0,
					PreAllocate:       0,
					MaxAboveWatermark: 0,
				},
			},
		},
		BrName:                    "br0",
		IsGW:                      false,
		IsDefaultGW:               false,
		ForceAddress:              false,
		IPMasq:                    true,
		IPMasqBackend:             &IPMasqBackend,
		MTU:                       1500,
		HairpinMode:               false,
		PromiscMode:               false,
		Vlan:                      0,
		VlanTrunk:                 nil,
		PreserveDefaultVlan:       false,
		MacSpoofChk:               false,
		EnableDad:                 false,
		DisableContainerInterface: false,
		PortIsolation:             false,
		Args: struct {
			Cni Args `json:"cni,omitempty" yaml:"cni,omitempty"`
		}{},
		RuntimeConfig: struct {
			Mac string `json:"mac,omitempty" yaml:"mac,omitempty"`
		}{},
		mac:   "",
		vlans: nil,
	}

	t.Run("TestBridgeByName", func(t *testing.T) {
		br, err := bridgeByName(h, "br0")
		if err != nil {
			t.Fatal("bridgeByName failed: ", err)
		}
		list, err := h.AddrStrList(br, netlink.FAMILY_ALL)
		if err != nil {
			t.Fatal(err)
		}
		msg := fmt.Sprintf("bridge %v", br)
		t.Log(msg)
		t.Log(list)
	})

	t.Run("TestSetupBridge", func(t *testing.T) {
		br, err := SetupBridge(h, c)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("bridge %v", br)
	})

	t.Run("TestRemoveBridge", func(t *testing.T) {
		err := RemoveBridge(h, c)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("TestRemoveBridgeIPAddr", func(t *testing.T) {
		ips := []string{
			"10.20.1.1/24",
			"10.20.3.1/24",
		}
		br, err := bridgeByName(h, "br0")
		if err != nil {
			t.Fatal("bridgeByName failed: ", err)
		}
		list, err := h.AddrStrList(br, netlink.FAMILY_V4)
		if err != nil {
			t.Fatal(err)
		}
		for _, nip := range ips {
			if slices.Contains(list, nip) {
				addr, err := netlink.ParseAddr(nip)
				if err != nil {
					t.Fatal(err)
				}
				err = h.AddrDel(br, addr)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}
