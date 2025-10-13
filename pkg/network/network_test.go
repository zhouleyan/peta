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
	"net"
	"testing"
)

func TestIPConfig(t *testing.T) {
	ipc := &IPConfig{
		Address: mustCIDR("10.1.1.111/24"),
		Gateway: net.ParseIP("10.1.1.1"),
	}
	data, err := ipc.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func mustCIDR(s string) net.IPNet {
	ip, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	n.IP = ip

	return *n
}
