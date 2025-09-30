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
	"errors"
	"fmt"
	"math/big"
	"net"
	"sync"
)

var (
	ErrFull              = errors.New("range is full")
	ErrAllocated         = errors.New("provided IP is already allocated")
	ErrMismatchedNetwork = errors.New("the provided network does not match the current range")
)

type ErrNotInRange struct {
	ValidRange string
}

func (e *ErrNotInRange) Error() string {
	return fmt.Sprintf("provided IP is not in the valid range. The range of valid IPs is %s", e.ValidRange)
}

// Range is a contiguous block of IPs that can be allocated atomically.
//
// The internal structure of the range is:
//
//	For CIDR 10.0.0.0/24
//	254 addresses usable out of 256 total (minus base and broadcast IPs)
//	  The number of usable addresses is r.max
//
//	CIDR base IP          CIDR broadcast IP
//	10.0.0.0                     10.0.0.255
//	|                                     |
//	0 1 2 3 4 5 ...         ... 253 254 255
//	  |                              |
//	r.base                     r.base + r.max
//	  |                              |
//	offset #0 of r.allocated   last offset of r.allocated
type Range struct {
	net *net.IPNet
	// base is a cached version of the start IP in the CIDR range as a *big.Int
	base *big.Int
	// max is the maximum size of the usable addresses in the range
	max int

	alloc Interface
}

// NewCIDRRange creates a Range over a net.IPNet, calling allocator.NewAllocationMap to construct
// the backing store. Returned Range excludes first (base) and last addresses (max) if provided cidr
// has more than 2 addresses.
func NewCIDRRange(cidr *net.IPNet) *Range {
	base := bigForIP(cidr.IP)
	size := RangeSize(cidr)

	// for any CIDR other than /32 or /128:
	if size > 2 {
		// don't use the network broadcast
		size = max(0, size-2)
		// don't use the network base
		base = base.Add(base, big.NewInt(1))
	}

	return &Range{
		net:   cidr,
		base:  base,
		max:   int(size),
		alloc: NewAllocationMap(int(size), cidr.String()),
	}
}

// Free returns the count of IP addresses left in the range.
func (r *Range) Free() int {
	return r.alloc.Free()
}

// Used returns the count of IP addresses used in the range.
func (r *Range) Used() int {
	return r.max - r.alloc.Free()
}

// CIDR returns the CIDR covered by the range.
func (r *Range) CIDR() net.IPNet {
	return *r.net
}

// Allocate attempts to reserve the provided IP. ErrNotInRange or
// ErrAllocated will be returned if the IP is not valid for this range
// or has already been reserved.  ErrFull will be returned if there
// are no addresses left.
func (r *Range) Allocate(ip net.IP) error {
	ok, offset := r.contains(ip)
	if !ok {
		return &ErrNotInRange{r.net.String()}
	}

	allocated := r.alloc.Allocate(offset)
	if !allocated {
		return ErrAllocated
	}
	return nil
}

// AllocateNext reserves one of the IPs from the pool. ErrFull may
// be returned if there are no addresses left.
func (r *Range) AllocateNext() (net.IP, error) {
	offset, ok := r.alloc.AllocateNext()
	if !ok {
		return nil, ErrFull
	}
	return addIPOffset(r.base, offset), nil
}

// Release releases the IP back to the pool. Releasing an
// unallocated IP or an IP out of the range is a no-op and
// returns no error.
func (r *Range) Release(ip net.IP) {
	ok, offset := r.contains(ip)
	if ok {
		r.alloc.Release(offset)
	}
}

// ForEach calls the provided function for each allocated IP.
func (r *Range) ForEach(fn func(net.IP)) {
	r.alloc.ForEach(func(offset int) {
		ip, _ := GetIndexedIP(r.net, offset+1) // +1 because Range doesn't store IP 0
		fn(ip)
	})
}

// Has returns true if the provided IP is already allocated and a call
// to Allocate(ip) would fail with ErrAllocated.
func (r *Range) Has(ip net.IP) bool {
	ok, offset := r.contains(ip)
	if !ok {
		return false
	}

	return r.alloc.Has(offset)
}

// Snapshot saves the current state of the pool.
func (r *Range) Snapshot() (string, []byte, error) {
	snapshottable, ok := r.alloc.(Snapshottable)
	if !ok {
		return "", nil, fmt.Errorf("not a snapshottable allocator")
	}
	str, data := snapshottable.Snapshot()
	return str, data, nil
}

// Restore restores the pool to the previously captured state. ErrMismatchedNetwork
// is returned if the provided IPNet range doesn't exactly match the previous range.
func (r *Range) Restore(net *net.IPNet, data []byte) error {
	if !net.IP.Equal(r.net.IP) || net.Mask.String() != r.net.Mask.String() {
		return ErrMismatchedNetwork
	}
	snapshottable, ok := r.alloc.(Snapshottable)
	if !ok {
		return fmt.Errorf("not a snapshottable allocator")
	}
	if err := snapshottable.Restore(net.String(), data); err != nil {
		return fmt.Errorf("restoring snapshot encountered: %w", err)
	}
	return nil
}

// contains returns true and the offset if the ip is in the range, and false
// and nil otherwise. The first and last addresses of the CIDR are omitted.
func (r *Range) contains(ip net.IP) (bool, int) {
	if !r.net.Contains(ip) {
		return false, 0
	}

	offset := calculateIPOffset(r.base, ip)
	if offset < 0 || offset >= r.max {
		return false, 0
	}
	return true, offset
}

// RangeSize returns the size of a range in valid addresses.
func RangeSize(subnet *net.IPNet) int64 {
	ones, bits := subnet.Mask.Size()
	if bits == 32 && (bits-ones) >= 31 || bits == 128 && (bits-ones) >= 127 {
		return 0
	}
	// For IPv6, the max size will be limited to 65536
	// This is due to the allocator keeping track of all the
	// allocated IP's in a bitmap. This will keep the size of
	// the bitmap to 64k.
	if bits == 128 && (bits-ones) >= 16 {
		return int64(1) << uint(16)
	} else {
		return int64(1) << uint(bits-ones)
	}
}

// GetIndexedIP returns a net.IP that is subnet.IP + index in the contiguous IP space.
func GetIndexedIP(subnet *net.IPNet, index int) (net.IP, error) {
	ip := addIPOffset(bigForIP(subnet.IP), index)
	if !subnet.Contains(ip) {
		return nil, fmt.Errorf("can't generate IP with index %d from subnet. subnet too small. subnet: %q", index, subnet)
	}
	return ip, nil
}

// bigForIP creates a big.Int based on the provided net.IP
func bigForIP(ip net.IP) *big.Int {
	// NOTE: Convert to 16-byte representation so we can
	// handle v4 and v6 values the same way.
	return big.NewInt(0).SetBytes(ip.To16())
}

// addIPOffset adds the provided integer offset to a base big.Int representing a net.IP
// NOTE: If you started with a v4 address and overflow it, you get a v6 result.
func addIPOffset(base *big.Int, offset int) net.IP {
	r := big.NewInt(0).Add(base, big.NewInt(int64(offset))).Bytes()
	r = append(make([]byte, 16), r...)
	return net.IP(r[len(r)-16:])
}

// calculateIPOffset calculates the integer offset of ip from base such that
// base + offset = ip. It requires ip >= base.
func calculateIPOffset(base *big.Int, ip net.IP) int {
	return int(big.NewInt(0).Sub(bigForIP(ip), base).Int64())
}

// AllocationResult is the result of an allocation
type AllocationResult struct {
	// IP is the allocated IP
	IP net.IP

	// IPPoolName is the IPAM pool from which the above IP was allocated from
	IPPoolName Pool

	// CIDRs is a list of all CIDRs to which the IP has direct access to.
	// This is primarily useful if the IP has been allocated out of a VPC
	// subnet range and the VPC provides routing to a set of CIDRs in which
	// the IP is routable
	CIDRs []string

	// PrimaryMAC is the MAC address of the primary interface. This is useful
	// when the IP is a secondary address of an interface which is
	// represented on the node as a Linux device and all routing of the IP
	// must occur through that master interface.
	PrimaryMAC string

	// GatewayIP is the IP of the gateway which must be used for this IP.
	// If the allocated IP is derived from a VPC, then the gateway
	// represented the gateway of the VPC or VPC subnet.
	GatewayIP string

	// ExpirationUUID is the UUID of the expiration timer. This field is
	// only set if AllocateNextWithExpiration is used.
	ExpirationUUID string

	// InterfaceNumber is a field for generically identifying an interface.
	// This is only useful in ENI mode.
	InterfaceNumber string
}

// Allocator is the interface for an IP allocator implementation
type Allocator interface {
	// Allocate allocates a specific IP or fails
	Allocate(ip net.IP, owner string, pool Pool) (*AllocationResult, error)

	// Release releases a previously allocated IP or fails
	Release(ip net.IP, pool Pool) error
}

type hostScopeAllocator struct {
	// mutex protects access to the allocated map
	mutex sync.RWMutex

	allocCIDR *net.IPNet
	allocator Allocator
}

func newHostScopeAllocator(family Family) *hostScopeAllocator {

	allocator := &hostScopeAllocator{}

	return allocator
}

// Allocate will attempt to find the specified IP in the custom resource and
// allocate it if it is available. If the IP is unavailable or already
// allocated, an error is returned. The custom resource will be updated to
// reflect the newly allocated IP.
func (a *hostScopeAllocator) Allocate(ip net.IP, owner string, pool Pool) (*AllocationResult, error) {
	return nil, nil
}
