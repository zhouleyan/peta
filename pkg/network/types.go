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
	"net"

	"peta.io/peta/pkg/network/ipam"
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

type IPAM struct {
	Type string `json:"type" yaml:"type"` // host-local

	ipam.Spec
}

// IsEmpty returns true if IPAM structure has no value, otherwise return false
func (i *IPAM) IsEmpty() bool {
	return i.Type == ""
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

// Conf describes a net configuration for a specific network.
type Conf struct {
	CNIVersion string `json:"cniVersion,omitempty"`

	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	Version      string `json:"version,omitempty" yaml:"version,omitempty"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	Capabilities []bool `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	IPAM         IPAM   `json:"ipam,omitempty" yaml:"ipam,omitempty"`
}
