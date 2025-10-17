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

package netlinksafe

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

// Arbitrary limit on max attempts at netlink calls if they are repeatedly interrupted.
const maxAttempts = 5

type Handle struct {
	*netlink.Handle
}

func NewHandle(nlFamilies ...int) (Handle, error) {
	nlh, err := netlink.NewHandle(nlFamilies...)
	if err != nil {
		return Handle{}, err
	}
	return Handle{Handle: nlh}, nil
}

func (h Handle) Close() {
	if h.Handle != nil {
		h.Handle.Close()
	}
}

func retryOnIntr(f func() error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := f(); !errors.Is(err, netlink.ErrDumpInterrupted) {
			return
		}
	}
	log.Printf("netlink call interrupted after %d attempts", maxAttempts)
}

func discardErrDumpInterrupted(err error) error {
	if errors.Is(err, netlink.ErrDumpInterrupted) {
		log.Printf("discarding ErrDumpInterrupted: %+v", errors.WithStack(err))
		return nil
	}
	return err
}

// LinkByName calls h.Handle.LinkByName, retrying if necessary. The netlink function
// doesn't normally ask the kernel for a dump of links. But, on an old kernel, it
// will do as a fallback and that dump may get inconsistent results.
func (h Handle) LinkByName(name string) (netlink.Link, error) {
	var link netlink.Link
	var err error
	retryOnIntr(func() error {
		link, err = h.Handle.LinkByName(name)
		return err
	})

	return link, discardErrDumpInterrupted(err)
}

// LinkList calls netlink.Handle.LinkList, retrying if necessary.
func LinkList() ([]netlink.Link, error) {
	var link []netlink.Link
	var err error
	retryOnIntr(func() error {
		link, err = netlink.LinkList()
		return err
	})
	return link, discardErrDumpInterrupted(err)
}

// LinkList calls h.Handle.LinkList, retrying if necessary.
func (h Handle) LinkList() ([]netlink.Link, error) {
	var links []netlink.Link
	var err error
	retryOnIntr(func() error {
		links, err = h.Handle.LinkList()
		return err
	})
	return links, discardErrDumpInterrupted(err)
}

func (h Handle) RouteListFiltered(family int, filter *netlink.Route, filterMask uint64) ([]netlink.Route, error) {
	var routes []netlink.Route
	var err error
	retryOnIntr(func() error {
		routes, err = h.Handle.RouteListFiltered(family, filter, filterMask)
		return err
	})
	return routes, discardErrDumpInterrupted(err)
}

// RouteList calls h.Handle.RouteList, retrying if necessary
func (h Handle) RouteList(link netlink.Link, family int) ([]netlink.Route, error) {
	var routes []netlink.Route
	var err error
	retryOnIntr(func() error {
		routes, err = h.Handle.RouteList(link, family)
		return err
	})
	return routes, discardErrDumpInterrupted(err)
}

// AddrList calls netlink.AddrList, retrying if necessary
func (h Handle) AddrList(link netlink.Link, family int) ([]netlink.Addr, error) {
	var addresses []netlink.Addr
	var err error
	retryOnIntr(func() error {
		addresses, err = h.Handle.AddrList(link, family)
		return err
	})
	return addresses, discardErrDumpInterrupted(err)
}

// AddrStrList calls AddrList, retrying if necessary
func (h Handle) AddrStrList(link netlink.Link, family int) ([]string, error) {
	var addrStrings []string
	var err error
	addresses, err := h.AddrList(link, family)
	for _, addr := range addresses {
		addrStrings = append(addrStrings, addr.IPNet.String())
	}
	return addrStrings, err
}

// BridgeVlanList calls netlink.BridgeVlanList, retrying if necessary
func (h Handle) BridgeVlanList() (map[int32][]*nl.BridgeVlanInfo, error) {
	var err error
	var info map[int32][]*nl.BridgeVlanInfo
	retryOnIntr(func() error {
		info, err = h.Handle.BridgeVlanList()
		return err
	})
	return info, discardErrDumpInterrupted(err)
}
