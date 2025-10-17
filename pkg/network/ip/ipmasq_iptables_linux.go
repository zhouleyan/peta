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
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"peta.io/peta/pkg/utils/iputils"
)

// setupIPMasqIPTables is the iptables-based implementation of SetupIPMasqForNetworks
func setupIPMasqIPTables(ipNs []*net.IPNet, network, _, id string) error {
	// Note: for historical reasons, the iptables implementation ignore ifName
	chain := iputils.FormatChainName(network, id)
	comment := iputils.FormatComment(network, id)
	for _, ipn := range ipNs {
		if err := setupIPMasq(ipn, chain, comment); err != nil {
			return err
		}
	}
	return nil
}

// SetupIPMasq installs iptables rules to masquerade traffic
// coming from ip if ipn and going outside ipn
func setupIPMasq(ipn *net.IPNet, chain, comment string) error {
	isV6 := ipn.IP.To4() == nil

	var ipt *iptables.IPTables
	var err error
	var multicastNet string

	if isV6 {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
		multicastNet = "ff00::/8"
	} else {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
		multicastNet = "224.0.0.0/4"
	}
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	// Create chan if it doesn't exist
	exists := false
	chains, err := ipt.ListChains("nat")
	if err != nil {
		return fmt.Errorf("failed to list nat chains: %v", err)
	}

	for _, ch := range chains {
		if ch == chain {
			exists = true
			break
		}
	}
	if !exists {
		if err := ipt.NewChain("nat", chain); err != nil {
			return fmt.Errorf("failed to create nat chain: %v", err)
		}
	}

	// Packets to this network should not be touched
	if err := ipt.AppendUnique("nat", chain, "-d", ipn.String(), "-j", "ACCEPT", "-m", "comment", "--comment", comment); err != nil {
		return err
	}

	// Don't masquerade multicast - pods should be able to talk to other pods
	// on the local network via multicast.
	if err := ipt.AppendUnique("nat", chain, "!", "-d", multicastNet, "-j", "MASQUERADE", "-m", "comment", "--comment", comment); err != nil {
		return err
	}

	// Packets from the specific IP of this network will hit the chain
	return ipt.AppendUnique("nat", "POSTROUTING", "-s", ipn.String(), "-j", chain, "-m", "comment", "--comment", comment)
}

// teardownIPMasqIPTables is the iptables-based implementation of TeardownIPMasqForNetworks
func teardownIPMasqIPTables(ipNs []*net.IPNet, network, _, id string) error {
	// Note: for historical reasons, the iptables implementation ignores ifName.
	chain := iputils.FormatChainName(network, id)
	comment := iputils.FormatComment(network, id)

	var errs []string
	for _, ipn := range ipNs {
		err := TeardownIPMasq(ipn, chain, comment)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if errs == nil {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}

// TeardownIPMasq undoes the effects of SetupIPMasq.
func TeardownIPMasq(ipn *net.IPNet, chain string, comment string) error {
	isV6 := ipn.IP.To4() == nil

	var ipt *iptables.IPTables
	var err error

	if isV6 {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv6)
	} else {
		ipt, err = iptables.NewWithProtocol(iptables.ProtocolIPv4)
	}
	if err != nil {
		return fmt.Errorf("failed to locate iptables: %v", err)
	}

	err = ipt.Delete("nat", "POSTROUTING", "-s", ipn.IP.String(), "-j", chain, "-m", "comment", "--comment", comment)
	if err != nil && !isNotExist(err) {
		return err
	}

	// for downward compatibility
	err = ipt.Delete("nat", "POSTROUTING", "-s", ipn.String(), "-j", chain, "-m", "comment", "--comment", comment)
	if err != nil && !isNotExist(err) {
		return err
	}

	err = ipt.ClearChain("nat", chain)
	if err != nil && !isNotExist(err) {
		return err
	}

	err = ipt.DeleteChain("nat", chain)
	if err != nil && !isNotExist(err) {
		return err
	}

	return nil
}

// isNotExist returns true if the error is from iptables indicating
// that the target does not exist.
func isNotExist(err error) bool {
	var e *iptables.Error
	ok := errors.As(err, &e)
	if !ok {
		return false
	}
	return e.IsNotExist()
}
