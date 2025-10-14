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
	"testing"

	"peta.io/peta/pkg/network/netlinksafe"
)

func TestVlan(t *testing.T) {
	t.Run("TestVlanCreate", func(t *testing.T) {
		h, err := netlinksafe.NewHandle()
		if err != nil {
			t.Fatal(err)
		}
		defer h.Close()
		brName := "virbr1"
		br, err := h.LinkByName(brName)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(br)
	})
}
