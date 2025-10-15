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

package network

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

// IPNet is like net.IPNet but adds JSON marshalling and unmarshalling
type IPNet net.IPNet

// ParseCIDR takes a string like "10.2.3.1/24" and
// return IPNet with "10.2.3.1" and /24 mask
func ParseCIDR(s string) (*net.IPNet, error) {
	ip, ipn, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	ipn.IP = ip
	return ipn, nil
}

// MarshalJSON
// 1. Only when the serialized object is a pointer type, will its pointer receiver's MarshalJSON method be called
// 2. If the object is a value type and does not have a MarshalJSON method with a value receiver,
// the default serialization logic will be used, and the method for a pointer receiver will not be called.
func (n IPNet) MarshalJSON() ([]byte, error) {
	return json.Marshal((*net.IPNet)(&n).String())
}

func (n *IPNet) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	tmp, err := ParseCIDR(s)
	if err != nil {
		return err
	}

	*n = IPNet(*tmp)
	return nil
}

// IPConfig contains values necessary to configure an IP address on an interface
type IPConfig struct {
	// Index into Result structs Interfaces list
	Interface *int
	Address   net.IPNet
	Gateway   net.IP
}

func (i *IPConfig) String() string {
	return fmt.Sprintf("%+v", *i)
}

func (i *IPConfig) Copy() *IPConfig {
	if i == nil {
		return nil
	}

	ipc := &IPConfig{
		Address: i.Address,
		Gateway: i.Gateway,
	}
	if i.Interface != nil {
		inf := *i.Interface
		ipc.Interface = &inf
	}
	return ipc
}

// JSON (un)marshall types
type ipConfig struct {
	Interface *int   `json:"interface,omitempty"`
	Address   IPNet  `json:"address"`
	Gateway   net.IP `json:"gateway,omitempty"`
}

func (i *IPConfig) MarshalJSON() ([]byte, error) {
	ipc := ipConfig{
		Interface: i.Interface,
		Address:   IPNet(i.Address),
		Gateway:   i.Gateway,
	}

	return json.Marshal(ipc)
}

func (i *IPConfig) UnmarshalJSON(data []byte) error {
	ipc := ipConfig{}
	if err := json.Unmarshal(data, &ipc); err != nil {
		return err
	}
	i.Interface = ipc.Interface
	i.Address = net.IPNet(ipc.Address)
	i.Gateway = ipc.Gateway
	return nil
}

// DNS contains values for DNS resolvers
type DNS struct {
	Nameservers []string `json:"nameservers" yaml:"nameservers"`
	Domain      string   `json:"domain" yaml:"domain"`
	Search      []string `json:"search" yaml:"search"`
	Options     []string `json:"options" yaml:"options"`
}

// IsEmpty returns true if DNS structure has no value, otherwise return false
func (d *DNS) IsEmpty() bool {
	if len(d.Nameservers) == 0 && d.Domain == "" && len(d.Search) == 0 && len(d.Options) == 0 {
		return true
	}
	return false
}

func (d *DNS) Copy() *DNS {
	if d == nil {
		return nil
	}

	to := &DNS{Domain: d.Domain}
	to.Nameservers = append(to.Nameservers, d.Nameservers...)
	to.Search = append(to.Search, d.Search...)
	to.Options = append(to.Options, d.Options...)
	return to
}

type Route struct {
	Dst      net.IPNet
	GW       net.IP
	MTU      int
	AdvMSS   int
	Priority int
	Table    *int
	Scope    *int
}

func (r *Route) String() string {
	table := "<nil>"
	if r.Table != nil {
		table = fmt.Sprintf("%d", *r.Table)
	}

	scope := "<nil>"
	if r.Scope != nil {
		scope = fmt.Sprintf("%d", *r.Scope)
	}

	return fmt.Sprintf("{Dst:%+v GW:%v MTU:%d AdvMSS:%d Priority:%d Table:%s Scope:%s}", r.Dst, r.GW, r.MTU, r.AdvMSS, r.Priority, table, scope)
}

func (r *Route) Copy() *Route {
	if r == nil {
		return nil
	}

	route := &Route{
		Dst:      r.Dst,
		GW:       r.GW,
		MTU:      r.MTU,
		AdvMSS:   r.AdvMSS,
		Priority: r.Priority,
		Scope:    r.Scope,
	}

	if r.Table != nil {
		table := *r.Table
		route.Table = &table
	}

	if r.Scope != nil {
		scope := *r.Scope
		route.Scope = &scope
	}

	return route
}

// net.IPNet is not JSON (un)marshall so this duality is needed
// for our custom IPNet type

// JSON (un)marshall types
type route struct {
	Dst      IPNet  `json:"dst"`
	GW       net.IP `json:"gw,omitempty"`
	MTU      int    `json:"mtu,omitempty"`
	AdvMSS   int    `json:"advmss,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Table    *int   `json:"table,omitempty"`
	Scope    *int   `json:"scope,omitempty"`
}

func (r *Route) UnmarshalJSON(data []byte) error {
	rt := route{}
	if err := json.Unmarshal(data, &rt); err != nil {
		return err
	}

	r.Dst = net.IPNet(rt.Dst)
	r.GW = rt.GW
	r.MTU = rt.MTU
	r.AdvMSS = rt.AdvMSS
	r.Priority = rt.Priority
	r.Table = rt.Table
	r.Scope = rt.Scope

	return nil
}

func (r Route) MarshalJSON() ([]byte, error) {
	rt := route{
		Dst:      IPNet(r.Dst),
		GW:       r.GW,
		MTU:      r.MTU,
		AdvMSS:   r.AdvMSS,
		Priority: r.Priority,
		Table:    r.Table,
		Scope:    r.Scope,
	}

	return json.Marshal(rt)
}

// IPAMSpec is the IPAM specification of the node
type IPAMSpec struct {
	// PodCIDRs is the list of CIDRs available to the node for allocation.
	// When an IP is used, the IP will be added to Used
	PodCIDR []string `json:"podCIDR" yaml:"podCIDR"`

	// MinAllocate is the minimum number of IPs that must be allocated when
	// the node is first bootstrapped.
	// It defines the minimum base socket of addresses that must be available.
	// After reaching this watermark, the PreAllocate and MaxAboveWatermark logic takes over to continue allocating IPs.
	MinAllocate int `json:"minAllocate" yaml:"minAllocate"`

	// MaxAllocate is the maximum number of IPs that can be allocated to the node.
	// When the current amount of allocated IPs will approach this value,
	// the considered value for PreAllocate will decrease down to 0 in order to
	// not attempt to allocate more addresses than defined.
	MaxAllocate int `json:"maxAllocate" yaml:"maxAllocate"`

	// PreAllocate defines the number of IP addresses that must be
	// available for allocation in the Spec.
	// It defines the buffer of
	// addresses available immediately without requiring cilium-operator to
	// get involved.
	PreAllocate int `json:"preAllocate" yaml:"preAllocate"`

	// MaxAboveWatermark is the maximum number of addresses to allocate
	// beyond the addresses needed to reach the PreAllocate watermark.
	// Going above the watermark can help reduce the number of API calls to
	// allocate IPs, e.g. when a new ENI is allocated, as many secondary
	// IPs as possible are allocated.
	// Limiting the amount can help reduce
	// waste of IPs.
	MaxAboveWatermark int `json:"maxAboveWatermark" yaml:"maxAboveWatermark"`
}

type IPAM struct {
	Type string `json:"type" yaml:"type"` // host-local

	IPAMSpec
}

// IsEmpty returns true if IPAM structure has no value, otherwise return false
func (i *IPAM) IsEmpty() bool {
	return i.Type == ""
}

// Conf describes a net configuration for a specific network.
type Conf struct {
	CNIVersion string `json:"cniVersion,omitempty" yaml:"cniVersion,omitempty"`

	Name         string          `json:"name,omitempty" yaml:"name,omitempty"`
	Type         string          `json:"type,omitempty" yaml:"type,omitempty"`
	Capabilities map[string]bool `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	IPAM         IPAM            `json:"ipam,omitempty" yaml:"ipam,omitempty"`
}

// Interface contains values about the created interfaces
type Interface struct {
	Name       string `json:"name"`
	Mac        string `json:"mac,omitempty"`
	Mtu        int    `json:"mtu,omitempty"`
	Sandbox    string `json:"sandbox,omitempty"`
	SocketPath string `json:"socketPath,omitempty"`
	PciID      string `json:"pciID,omitempty"`
}

func (i *Interface) String() string {
	return fmt.Sprintf("%+v", *i)
}

func (i *Interface) Copy() *Interface {
	if i == nil {
		return nil
	}
	newI := *i
	return &newI
}

type Result struct {
	CNIVersion string       `json:"cniVersion,omitempty" yaml:"cniVersion,omitempty"`
	Interfaces []*Interface `json:"interfaces,omitempty" yaml:"interfaces,omitempty"`
	IPs        []*IPConfig  `json:"ips,omitempty" yaml:"ips,omitempty"`
	Routes     []*Route     `json:"routes,omitempty" yaml:"routes,omitempty"`
	DNS        DNS          `json:"dns,omitempty" yaml:"dns,omitempty"`
}

// MarshalJSON Note: DNS should be omitted if DNS is empty but default Marshal function
// will output empty structure hence need to write a Marshal function
func (r *Result) MarshalJSON() ([]byte, error) {
	// use type alias to escape recursion for json.Marshal() to MarshalJSON()
	type fixObjType = Result

	bytes, err := json.Marshal(fixObjType(*r)) //nolint:all
	if err != nil {
		return nil, err
	}

	fixupObj := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &fixupObj); err != nil {
		return nil, err
	}

	if r.DNS.IsEmpty() {
		delete(fixupObj, "dns")
	}

	return json.Marshal(fixupObj)
}

func (r *Result) Version() string {
	return r.CNIVersion
}

func (r *Result) Print() error {
	return r.PrintTo(os.Stdout)
}

func (r *Result) PrintTo(writer io.Writer) error {
	data, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}
