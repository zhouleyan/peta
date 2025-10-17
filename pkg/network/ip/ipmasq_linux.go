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

	"peta.io/peta/pkg/utils/iputils"
)

// SetupIPMasqForNetworks installs rules to masquerade traffic coming from ips of ipNets and
// going outside ipNets, using a chain name based on network and ifName.
// The backend can be either "iptables" or "nftables"; if it is nil, then a suitable default
// implementation will be used.
func SetupIPMasqForNetworks(backend *string, ipNs []*net.IPNet, network, ifName, id string) error {
	if backend == nil {
		// Prefer iptables, unless only nftables is available
		defaultBackend := "iptables"
		if !iputils.SupportsIPTables() && iputils.SupportsNFTables() {
			defaultBackend = "nftables"
		}
		backend = &defaultBackend
	}
	switch *backend {
	case "iptables":
		return setupIPMasqIPTables(ipNs, network, ifName, id)
	case "nftables":
		return setupIPMasqNFTables(ipNs, network, ifName, id)
	default:
		return fmt.Errorf("unsupported ip backend: %s", *backend)
	}
}

// TeardownIPMasqForNetworks undoes the effects of SetupIPMasqForNetworks
func TeardownIPMasqForNetworks(ipNs []*net.IPNet, network, ifName, id string) error {
	var errs []string

	// Do both the iptables and the nftables cleanup, since the pod may have been
	// created with a different version of this plugin or a different configuration.

	err := teardownIPMasqIPTables(ipNs, network, ifName, id)
	if err != nil && iputils.SupportsIPTables() {
		errs = append(errs, err.Error())
	}

	err = teardownIPMasqNFTables(ipNs, network, ifName, id)
	if err != nil && iputils.SupportsNFTables() {
		errs = append(errs, err.Error())
	}

	if errs == nil {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}
