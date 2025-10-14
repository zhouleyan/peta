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

package netlinksafe

import (
	"net"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestLink(t *testing.T) {
	h, err := NewHandle()
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	t.Run("TestLinkByName", func(t *testing.T) {
		link, err := h.LinkByName("virbr0")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(link)
	})

	t.Run("TestLinkList", func(t *testing.T) {
		list, err := h.LinkList()
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range list {
			t.Log(l.Attrs().Name)
		}
	})

	t.Run("TestRouteListFiltered", func(t *testing.T) {
		var err error
		_, prefix, err := net.ParseCIDR("10.1.1.0/24")
		if err != nil {
			t.Fatal(err)
		}
		findGwy := &netlink.Route{Dst: prefix}
		routes, err := h.RouteListFiltered(netlink.FAMILY_V4,
			findGwy,
			netlink.RT_FILTER_DST,
		)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range routes {
			t.Log(r)
		}
	})

	t.Run("TestRouteList", func(t *testing.T) {
		routes, err := h.RouteList(nil, netlink.FAMILY_ALL)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range routes {
			t.Log(r)
		}
	})

	t.Run("TestVlanList", func(t *testing.T) {
		l, err := h.BridgeVlanList()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(l)
	})
}
