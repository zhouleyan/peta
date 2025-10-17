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
	"context"
	"fmt"
	"net"
	"strings"

	"peta.io/peta/pkg/utils/iputils"
	"sigs.k8s.io/knftables"
)

const (
	ipMasqTableName = "cni_plugins_masquerade"
	ipMasqChainName = "masq_checks"
)

// The nftables ip masq implementation is mostly like the iptables implementation, with
// minor updates to fix a bug (adding `ifName`) and to allow future GC support.
//
// We add a rule for each mapping, with a comment containing a hash of its identifiers,
// so that we can later reliably delete the rules we want. (This is important because in
// edge cases, it's possible the plugin might see "ADD container A with IP 192.168.1.3",
// followed by "ADD container B with IP 192.168.1.3" followed by "DEL container A with IP
// 192.168.1.3", and we need to make sure that the DEL causes us to delete the rule for
// container A, and not the rule for container B.)
//
// It would be more nftables-y to have a chain with a single rule doing a lookup against a
// set with an element per mapping, rather than having a chain with a rule per mapping.
// But there's no easy, non-racy way to say "delete the element 192.168.1.3 from the set,
// but only if it was added for container A, not if it was added for container B".

// hashForNetwork returns a unique hash for this network
func hashForNetwork(network string) string {
	return iputils.MustFormatHashWithPrefix(16, "", network)
}

// hashForInstance returns a unique hash identifying the rules for this
// network/ifName/containerID
func hashForInstance(network, ifName, id string) string {
	return hashForNetwork(network) + "-" + iputils.MustFormatHashWithPrefix(16, "", ifName+":"+id)
}

// commentForInstance returns a comment string that begins with a unique hash and
// ends with a (possibly-truncated) human-readable description.
func commentForInstance(network, ifName, id string) string {
	comment := fmt.Sprintf("%s, net: %s, if: %s, id: %s",
		hashForInstance(network, ifName, id),
		strings.ReplaceAll(network, `"`, ``),
		strings.ReplaceAll(ifName, `"`, ``),
		strings.ReplaceAll(id, `"`, ``),
	)
	if len(comment) > knftables.CommentLengthMax {
		comment = comment[:knftables.CommentLengthMax]
	}
	return comment
}

// setupIPMasqNFTables is the nftables-based implementation of SetupIPMasqForNetworks
func setupIPMasqNFTables(ipNs []*net.IPNet, network, ifName, id string) error {
	nft, err := knftables.New(knftables.InetFamily, ipMasqTableName)
	if err != nil {
		return err
	}
	return setupIPMasqNFTablesWithInterface(nft, ipNs, network, ifName, id)
}

func setupIPMasqNFTablesWithInterface(nft knftables.Interface, ipNs []*net.IPNet, network, ifName, id string) error {
	staleRules, err := findRules(nft, hashForInstance(network, ifName, id))
	if err != nil {
		return err
	}

	tx := nft.NewTransaction()

	// Ensure that our table and chains exist.
	tx.Add(&knftables.Table{
		Comment: knftables.PtrTo("Masquerading for plugins from github.com/containernetworking/plugins"),
	})
	tx.Add(&knftables.Chain{
		Name:    ipMasqChainName,
		Comment: knftables.PtrTo("Masquerade traffic from certain IPs to any (non-multicast) IP outside their subnet"),
	})

	// Ensure that the postrouting chain exists and has the correct rules. (Has to be
	// done after creating ipMasqChainName, so we can jump to it.)
	tx.Add(&knftables.Chain{
		Name:     "postrouting",
		Type:     knftables.PtrTo(knftables.NATType),
		Hook:     knftables.PtrTo(knftables.PostroutingHook),
		Priority: knftables.PtrTo(knftables.SNATPriority),
	})
	tx.Flush(&knftables.Chain{
		Name: "postrouting",
	})
	tx.Add(&knftables.Rule{
		Chain: "postrouting",
		Rule:  "ip daddr == 224.0.0.0/4  return",
	})
	tx.Add(&knftables.Rule{
		Chain: "postrouting",
		Rule:  "ip6 daddr == ff00::/8  return",
	})
	tx.Add(&knftables.Rule{
		Chain: "postrouting",
		Rule: knftables.Concat(
			"goto", ipMasqChainName,
		),
	})

	// Delete stale rules, add new rules to masquerade chain
	for _, rule := range staleRules {
		tx.Delete(rule)
	}
	for _, ipn := range ipNs {
		ip := "ip"
		if ipn.IP.To4() == nil {
			ip = "ip6"
		}

		// e.g. if ipn is "192.168.1.4/24", then dstNet is "192.168.1.0/24"
		dstNet := &net.IPNet{IP: ipn.IP.Mask(ipn.Mask), Mask: ipn.Mask}

		tx.Add(&knftables.Rule{
			Chain: ipMasqChainName,
			Rule: knftables.Concat(
				ip, "saddr", "==", ipn.IP,
				ip, "daddr", "!=", dstNet,
				"masquerade",
			),
			Comment: knftables.PtrTo(commentForInstance(network, ifName, id)),
		})
	}

	return nft.Run(context.TODO(), tx)
}

// findRules finds rules with comments that start with commentPrefix.
func findRules(nft knftables.Interface, commentPrefix string) ([]*knftables.Rule, error) {
	rules, err := nft.ListRules(context.TODO(), ipMasqChainName)
	if err != nil {
		if knftables.IsNotFound(err) {
			// If ipMasqChainName doesn't exist yet, that's fine
			return nil, nil
		}
		return nil, err
	}

	matchingRules := make([]*knftables.Rule, 0, 1)
	for _, rule := range rules {
		if rule.Comment != nil && strings.HasPrefix(*rule.Comment, commentPrefix) {
			matchingRules = append(matchingRules, rule)
		}
	}

	return matchingRules, nil
}

// teardownIPMasqNFTables is the nftables-based implementation of TeardownIPMasqForNetworks
func teardownIPMasqNFTables(ipNs []*net.IPNet, network, ifName, id string) error {
	nft, err := knftables.New(knftables.InetFamily, ipMasqTableName)
	if err != nil {
		return err
	}
	return teardownIPMasqNFTablesWithInterface(nft, ipNs, network, ifName, id)
}

func teardownIPMasqNFTablesWithInterface(nft knftables.Interface, _ []*net.IPNet, network, ifName, id string) error {
	rules, err := findRules(nft, hashForInstance(network, ifName, id))
	if err != nil {
		return err
	} else if len(rules) == 0 {
		return nil
	}

	tx := nft.NewTransaction()
	for _, rule := range rules {
		tx.Delete(rule)
	}
	return nft.Run(context.TODO(), tx)
}
