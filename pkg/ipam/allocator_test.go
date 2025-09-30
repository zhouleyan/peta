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

package ipam

import (
	"fmt"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustParseCidr(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(fmt.Errorf("net.ParseCIDR: %w", err))
	}
	return ipNet
}

func ipForBig(i *big.Int) net.IP {
	return addIPOffset(i, 0)
}

func TestNewCIDRRange(t *testing.T) {
	testCases := []struct {
		name     string
		ipNet    *net.IPNet
		wantBase net.IP
		wantMax  int
		wantSize int64
	}{
		{
			name:     "IPv4 /24",
			ipNet:    mustParseCidr("192.168.0.1/24"),
			wantBase: net.ParseIP("192.168.0.1"),
			wantMax:  254, // (2^(32-27) - 2
			wantSize: 256,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewCIDRRange(tc.ipNet)
			ip, err := actual.AllocateNext()
			if err != nil {
				t.Errorf("NewCIDRRange() error = %v", err)
			}
			t.Log(ip.String())

			err = actual.Allocate(net.ParseIP("192.168.0.160"))
			if err != nil {
				t.Errorf("Allocate() error = %v", err)
			}
			t.Log(actual.Has(net.ParseIP("192.168.0.160")))
			t.Log(actual.Free())
			actual.ForEach(func(ip net.IP) {
				t.Log(ip.String())
			})
			t.Log(actual.Free())
			t.Log(actual.Used())
			name, bytes, err := actual.Snapshot()
			if err != nil {
				t.Errorf("Snapshot() error = %v", err)
			}
			data := big.NewInt(0).SetBytes(bytes)
			t.Log(name)
			t.Log(data)
			baseIP := ipForBig(actual.base)
			require.Equal(t, tc.wantBase.String(), baseIP.String())
			require.Equal(t, tc.wantMax, actual.max)
			require.Equal(t, tc.wantSize, RangeSize(tc.ipNet))
		})
	}
}
