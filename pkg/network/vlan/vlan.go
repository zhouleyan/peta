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

package vlan

import (
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"
	"peta.io/peta/pkg/network/ip"
	"peta.io/peta/pkg/network/netlinksafe"
)

type Conf struct {
	Master string `json:"master"`
	VlanID int    `json:"vlanId"`
	MTU    int    `json:"mtu,omitempty"`
}

func init() {
	runtime.LockOSThread()
}

func CreateVlan(h netlinksafe.Handle, ifName string, c *Conf) (*netlink.Vlan, error) {
	var m netlink.Link
	var err error

	if c.Master == "" {
		return nil, fmt.Errorf("\"master\" field is required. It specifies the host interface name to create the VLAN for")
	}

	if c.VlanID < 0 || c.VlanID > 4094 {
		return nil, fmt.Errorf("invalid VLAN ID %d (must be between 0 and 4095 inclusive)", c.VlanID)
	}

	// check existing and MTU of master interface
	masterMTU, err := getMTUByName(h, c.Master)
	if err != nil {
		return nil, err
	}
	if c.MTU < 0 || c.MTU > masterMTU {
		return nil, fmt.Errorf("invalid MTU %d, must be [0, master MTU(%d)]", c.MTU, masterMTU)
	}

	m, err = h.LinkByName(c.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", c.Master, err)
	}

	// due to kernel bug we have to create with tmpName, or it might
	// collide with the name on the host and error out
	//tmpName, err := ip.RandomVethName()
	//if err != nil {
	//	return nil, err
	//}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.MTU = c.MTU
	linkAttrs.Name = ifName
	linkAttrs.ParentIndex = m.Attrs().Index

	v := &netlink.Vlan{
		LinkAttrs: linkAttrs,
		VlanId:    c.VlanID,
	}

	err = h.LinkAdd(v)
	if err != nil {
		return nil, fmt.Errorf("failed to create vlan: %v", err)
	}

	return v, err
}

func DeleteVlan(h netlinksafe.Handle, ifName string) error {
	return ip.DelLinkByName(h, ifName)
}

func getMTUByName(h netlinksafe.Handle, ifName string) (int, error) {
	var link netlink.Link
	var err error
	link, err = h.LinkByName(ifName)
	if err != nil {
		return 0, err
	}
	return link.Attrs().MTU, nil
}
