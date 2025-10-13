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

	"github.com/vishvananda/netlink"
	"peta.io/peta/pkg/network"
	"peta.io/peta/pkg/network/netlinksafe"
)

const defaultBrName = "br0"

func CreateBridge(h netlinksafe.Handle, conf *network.BrConf) (*netlink.Bridge, error) {
	// 1. Check config
	isLayer3 := conf.IPAM.Type != ""

	if isLayer3 && conf.DisableContainerInterface {
		return nil, fmt.Errorf("cannot use IPAM when DisableContainerInterface flag is set")
	}

	if conf.IsDefaultGW {
		conf.IsGW = true
	}

	if conf.HairpinMode && conf.PromiscMode {
		return nil, fmt.Errorf("cannot set hairpin mode and promiscuous mode at the same time")
	}

	// 2. Enable IP forward

	// 3. Create bridge

	// 4. Set bridge IP

	// 5. Set bridge UP

	// 6. Set IPAM

	// 7. Set NAT

	return nil, nil
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
