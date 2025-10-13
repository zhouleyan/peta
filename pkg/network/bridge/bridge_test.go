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
	"testing"

	"peta.io/peta/pkg/network/netlinksafe"
)

func TestBridge(t *testing.T) {
	h, err := netlinksafe.NewHandle()
	if &h == nil {
		t.Fatal("NewHandle failed")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	t.Run("bridgeByName", func(t *testing.T) {
		br, err := bridgeByName(h, "ens33")
		if err != nil {
			t.Fatal("bridgeByName failed: ", err)
		}
		msg := fmt.Sprintf("bridge %v", br)
		t.Log(msg)
	})
}
