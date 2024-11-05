/*
 *  This file is part of PETA.
 *  Copyright (C) 2024 The PETA Authors.
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

package request

import (
	"context"
	"fmt"
	"net/http"
	"peta.io/peta/pkg/apis"
	"peta.io/peta/pkg/utils/iputils"
	"strings"
)

const (
	VerbCreate = "create"
	VerbGet    = "get"
	VerbList   = "list"
	VerbUpdate = "update"
	VerbDelete = "delete"
	VerbWatch  = "watch"
	VerbPatch  = "patch"
)

type InfoResolver interface {
	NewRequestInfo(req *http.Request) (*Info, error)
}

type requestInfoKeyType int

// requestInfoKey is the Info key for the context. It's of private type here. Because
// keys are interfaces and interfaces are equal when the type and the value is equal, this
// does not conflict with the keys defined in pkg/api.
const requestInfoKey requestInfoKeyType = iota

// Info holds information parsed from the http.Request
type Info struct {
	// IsResourceRequest indicates whether the request is for an API resource or subresource
	IsResourceRequest bool

	// Path is the URL path of the request
	Path string

	// Verb is the peta verb associated with the request for API requests, not the http verb. This includes things like
	// list and watch.
	// For non-resource requests, this is the lowercase http verb
	Verb string

	APIPrefix string

	APIGroup string

	APIVersion string

	// Cluster of requested resource, this is empty in single-cluster environment
	Cluster string

	Workspace string

	Namespace string

	// Resource is the name of the resource being requested. This is not the kind.
	Resource string

	// Scope of requested resource
	ResourceScope string

	// Subresource is the name of the subresource being requested.
	// For instance, /users has the resource "users" and the kind "User", while /users/foo/status has the resource "users",
	// the subresource "status", and the kind "User"
	Subresource string

	// Name is empty for some verb, but if the request directly indicates a name (not in body content) then this field is filled in.
	Name string

	// Parts are the path parts for the request, always starting with /{resource}/{name}
	Parts []string

	// Source IP
	SourceIP string

	// User agent
	UserAgent string
}

type InfoFactory struct {
}

// NewRequestInfo returns the information from the http request.
func (i *InfoFactory) NewRequestInfo(req *http.Request) (*Info, error) {
	info := Info{
		Path:      req.URL.Path,
		Verb:      req.Method,
		Workspace: apis.WorkspaceNone,
		Cluster:   apis.ClusterNone,
		SourceIP:  iputils.RemoteIP(req),
		UserAgent: req.UserAgent(),
	}

	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &info, nil
	}

	// URL forms: /clusters/{cluster}/*
	if currentParts[0] == "clusters" {
		if len(currentParts) > 1 {
			info.Cluster = currentParts[1]
			info.Path = strings.TrimPrefix(info.Path, fmt.Sprintf("/clusters/%s", info.Cluster))
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	return &info, nil
}

func WithRequestInfo(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, requestInfoKey, info)
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
