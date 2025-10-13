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

package ip

import (
	"bytes"
	"os"

	"peta.io/peta/pkg/network"
)

func EnableIP4Forward() error {
	return echo1("/proc/sys/net/ipv4/ip_forward")
}

func EnableIP6Forward() error {
	return echo1("/proc/sys/net/ipv6/conf/all/forwarding")
}

func EnableForward(ips []*network.IPConfig) error {
	v4 := false
	v6 := false

	for _, ip := range ips {
		isV4 := ip.Address.IP.To4() != nil
		if isV4 && !v4 {
			if err := EnableIP4Forward(); err != nil {
				return err
			}
			v4 = true
		} else if !isV4 && !v6 {
			if err := EnableIP6Forward(); err != nil {
				return err
			}
			v6 = true
		}
	}
	return nil
}

func echo1(f string) error {
	if content, err := os.ReadFile(f); err != nil {
		if bytes.Equal(bytes.TrimSpace(content), []byte("1")) {
			return nil
		}
	}
	return os.WriteFile(f, []byte("1"), 0o644)
}
