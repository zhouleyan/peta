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

import "net"

// Family is the type describing all address families support by the IP
// allocation manager
type Family string

const (
	IPv4 Family = "ipv4"
	IPv6 Family = "ipv6"
)

type IPAM struct {
	Type string `json:"type" yaml:"type"` // host-local

	Spec
}

// IsEmpty returns true if IPAM structure has no value, otherwise return false
func (i *IPAM) IsEmpty() bool {
	return i.Type == ""
}

// Spec is the IPAM specification of the node
type Spec struct {
	// PodCIDRs is the list of CIDRs available to the node for allocation.
	// When an IP is used, the IP will be added to Used
	PodCIDR []string `json:"podCIDR" yaml:"podCIDR"`

	// MinAllocate is the minimum number of IPs that must be allocated when
	// the node is first bootstrapped.
	// It defines the minimum base socket of addresses that must be available.
	// After reaching this watermark, the PreAllocate and MaxAboveWatermark logic takes over to continue allocating IPs.
	MinAllocate int `json:"minAllocate" yaml:"minAllocate"`

	// MaxAllocate is the maximum number of IPs that can be allocated to the node.
	// When the current amount of allocated IPs will approach this value,
	// the considered value for PreAllocate will decrease down to 0 in order to
	// not attempt to allocate more addresses than defined.
	MaxAllocate int `json:"maxAllocate" yaml:"maxAllocate"`

	// PreAllocate defines the number of IP addresses that must be
	// available for allocation in the Spec.
	// It defines the buffer of
	// addresses available immediately without requiring cilium-operator to
	// get involved.
	PreAllocate int `json:"preAllocate" yaml:"preAllocate"`

	// MaxAboveWatermark is the maximum number of addresses to allocate
	// beyond the addresses needed to reach the PreAllocate watermark.
	// Going above the watermark can help reduce the number of API calls to
	// allocate IPs, e.g. when a new ENI is allocated, as many secondary
	// IPs as possible are allocated.
	// Limiting the amount can help reduce
	// waste of IPs.
	MaxAboveWatermark int `json:"maxAboveWatermark" yaml:"maxAboveWatermark"`
}

// AllocationIP is an IP which is available for allocation, or already
// has been allocated
type AllocationIP struct {
	// Owner is the owner of the IP. This field is set if the IP has been
	// allocated. It will be set to the pod name or another identifier
	// representing the usage of the IP
	//
	// The owner field is left blank for an entry in Spec.IPAM.Pool and
	// filled out as the IP is used and also added to Status.IPAM.Used.
	//
	// +optional
	Owner string `json:"owner,omitempty"`

	// Resource is set for both available and allocated IPs, it represents
	// what resource the IP is associated with, e.g. in combination with
	// AWS ENI, this will refer to the ID of the ENI
	//
	// +optional
	Resource string `json:"resource,omitempty"`
}

// AllocationMap is a map of allocated IPs indexed by IP
type AllocationMap map[string]AllocationIP

// DeriveFamily derives the address family of an IP
func DeriveFamily(ip net.IP) Family {
	if ip.To4() == nil {
		return IPv6
	}
	return IPv4
}

type Pool string

func (p Pool) String() string {
	return string(p)
}
