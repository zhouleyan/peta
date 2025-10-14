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
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"

	"github.com/vishvananda/netlink"
	"peta.io/peta/pkg/network/netlinksafe"
)

var ErrLinkNotFound = errors.New("link not found")

// CountIPsInCIDR takes a RFC4632/RFC4291-formatted IPv4/IPv6 CIDR and
// determines how many IP addresses reside within that CIDR.
// The first and the last (base and broadcast) IPs are excluded.
//
// Returns 0 if the input CIDR cannot be parsed.
func CountIPsInCIDR(ipNet *net.IPNet) *big.Int {
	subnet, size := ipNet.Mask.Size()
	if subnet == size {
		return big.NewInt(0)
	}
	return big.NewInt(0).
		Sub(
			big.NewInt(2).Exp(big.NewInt(2),
				big.NewInt(int64(size-subnet)), nil),
			big.NewInt(2),
		)
}

// DelLinkByName removes an interface link
func DelLinkByName(h netlinksafe.Handle, ifName string) error {
	l, err := h.LinkByName(ifName)
	if err != nil {
		var linkNotFoundError netlink.LinkNotFoundError
		if errors.As(err, &linkNotFoundError) {
			return ErrLinkNotFound
		}
		return fmt.Errorf("failed to lookup %q: %v", ifName, err)
	}

	if err = h.LinkDel(l); err != nil {
		return fmt.Errorf("failed to delete link %q: %v", l.Attrs().Name, err)
	}
	return nil
}

// DelLinkByNameAddr remove an interface and returns its addresses
func DelLinkByNameAddr(h netlinksafe.Handle, ifName string) ([]*net.IPNet, error) {
	l, err := h.LinkByName(ifName)
	if err != nil {
		var linkNotFoundError netlink.LinkNotFoundError
		if errors.As(err, &linkNotFoundError) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to lookup %q: %v", ifName, err)
	}

	addresses, err := h.AddrList(l, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP addresses for %q: %v", ifName, err)
	}

	if err = h.LinkDel(l); err != nil {
		return nil, fmt.Errorf("failed to delete link %q: %v", l.Attrs().Name, err)
	}

	var out []*net.IPNet
	for _, addr := range addresses {
		if addr.IP.IsGlobalUnicast() {
			out = append(out, addr.IPNet)
		}
	}

	return out, nil
}

// RandomVethName returns string "veth" with random prefix (hashed from entropy)
func RandomVethName() (string, error) {
	entropy := make([]byte, 4)
	_, err := rand.Read(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate random veth name: %v", err)
	}

	// NetworkManager (recent versions) will ignore veth devices that start with "veth"
	return fmt.Sprintf("veth%x", entropy), nil
}
