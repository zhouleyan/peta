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

package iputils

import (
	"crypto/sha512"
	"fmt"

	"github.com/coreos/go-iptables/iptables"
	"sigs.k8s.io/knftables"
)

const (
	maxChainLen = 28
	chainPrefix = "PETA-"
	MaxHashLen  = sha512.Size * 2
)

// FormatChainName generates a chain name to be used
// with iptables. Ensures that the generated chain
// name is exactly maxChainLength chars in length
func FormatChainName(name string, id string) string {
	return MustFormatChainNameWithPrefix(name, id, "")
}

// FormatComment returns a comment used for easier
// rule identification within iptables
func FormatComment(name, id string) string {
	return fmt.Sprintf("name %q id: %q", name, id)
}

// MustFormatChainNameWithPrefix generates a chain name similar
// to FormatChainName, but adds a custom prefix between
// chainPrefix and unique identifier. Ensures that the
// generated chain name is exactly maxChainLength chars in length.
// Panics if the given prefix is too long
func MustFormatChainNameWithPrefix(name string, id string, prefix string) string {
	return MustFormatHashWithPrefix(maxChainLen, chainPrefix+prefix, name+id)
}

// MustFormatHashWithPrefix returns a string of given length that begins with the
// given prefix. It is filled with entropy based on the given string toHash
func MustFormatHashWithPrefix(length int, prefix string, toHash string) string {
	if len(prefix) >= length || length > MaxHashLen {
		panic("invalid length")
	}

	output := sha512.Sum512([]byte(toHash))
	return fmt.Sprintf("%s%x", prefix, output)[:length]
}

// SupportsIPTables tests whether the system supports using netfilter via the iptables API
// (whether via "iptables-legacy" or "iptables-nft"). (Note that this returns true if it
// is *possible* to use iptables; it does not test whether any other components on the
// system are *actually* using iptables.)
func SupportsIPTables() bool {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return false
	}
	// We don't care whether the chain actually exists, only whether we can *check*
	// whether it exists.
	_, err = ipt.ChainExists("filter", "INPUT")
	return err == nil
}

// SupportsNFTables tests whether the system supports using netfilter via the nftables API
// (ie, not via "iptables-nft"). (Note that this returns true if it is *possible* to use
// nftables; it does not test whether any other components on the system are *actually*
// using nftables.)
func SupportsNFTables() bool {
	// knftables.New() does sanity checks so we don't need any further test like in
	// the iptables case.
	_, err := knftables.New(knftables.IPv4Family, "supports_nftables_test")
	return err == nil
}
