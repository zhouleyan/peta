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
	"peta.io/peta/pkg/utils/sets"
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

// specialVerbs contains just strings which are used in REST paths for special actions that don't fall under the normal
// CRUDdy GET/POST/PUT/DELETE actions on REST objects.
// master's Mux.
var specialVerbs = sets.New("proxy", "watch")

// namespaceSubResources contains subresource of namespace
// this list allows the parser to distinguish between a namespace subresource, and a namespaced resource
var namespaceSubResources = sets.New("status", "finalize")

// specialVerbsNoSubResources contains root verbs which do not allow subresource
var specialVerbsNoSubResources = sets.New("proxy")

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
	APIPrefixes sets.Set[string]
}

// NewRequestInfo returns the information from the http request.
func (ri *InfoFactory) NewRequestInfo(req *http.Request) (*Info, error) {
	info := Info{
		Path:      req.URL.Path,
		Verb:      req.Method,
		Workspace: apis.WorkspaceNone,
		Cluster:   apis.ClusterNone,
		SourceIP:  iputils.RemoteIP(req),
		UserAgent: req.UserAgent(),
	}

	// p0: "apis"
	// p1: "version.peta.io"
	// p2: "v1alpha1"
	// p3: "version"
	// ["apis", "version.peta.io", "v1alpha1", "version"]
	currentParts := splitPath(req.URL.Path)
	if len(currentParts) < 3 {
		return &info, nil
	}

	// URL forms: /clusters/{cluster}/*
	// Path: /clusters/foo/apis/version.peta.io/v1alpha1/version => /apis/version.peta.io/v1alpha1/version
	// Cluster: foo
	if currentParts[0] == "clusters" {
		if len(currentParts) > 1 {
			info.Cluster = currentParts[1]
			info.Path = strings.TrimPrefix(info.Path, fmt.Sprintf("/clusters/%s", info.Cluster))
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	if !ri.APIPrefixes.Has(currentParts[0]) {
		// return a non-resource request
		return &info, nil
	}
	info.APIPrefix = currentParts[0]
	// ["version.peta.io", "v1alpha1", "version"]
	currentParts = currentParts[1:]

	if len(currentParts) < 3 {
		return &info, nil
	}

	info.APIGroup = currentParts[0]
	// ["v1alpha1", "version"]
	currentParts = currentParts[1:]

	info.IsResourceRequest = true
	info.APIVersion = currentParts[0]
	// ["version"]
	currentParts = currentParts[1:]

	if len(currentParts) > 0 && specialVerbs.Has(currentParts[0]) {
		if len(currentParts) < 2 {
			return &info, fmt.Errorf("unable to determine kind and namespace from url: %v", req.URL)
		}

		info.Verb = currentParts[0]
		currentParts = currentParts[1:]
	} else {
		switch req.Method {
		case "POST":
			info.Verb = VerbCreate
		case "GET", "HEAD":
			info.Verb = VerbGet
		case "PUT":
			info.Verb = VerbUpdate
		case "PATCH":
			info.Verb = VerbPatch
		case "DELETE":
			info.Verb = VerbDelete
		default:
			info.Verb = ""
		}
	}

	// URL forms: /workspaces/{workspace}/*
	if currentParts[0] == "workspaces" {
		if len(currentParts) > 1 {
			info.Workspace = currentParts[1]
		}
		if len(currentParts) > 2 {
			currentParts = currentParts[2:]
		}
	}

	// URL forms: /namespaces/{namespace}/{kind}/*, where parts are adjusted to be relative to kind
	if currentParts[0] == "namespaces" {
		if len(currentParts) > 1 {
			info.Namespace = currentParts[1]

			// if there is another step after the namespace name, and it is not a known namespace subresource
			// move currentParts to include it as a resource in its own right
			if len(currentParts) > 2 && !namespaceSubResources.Has(currentParts[2]) {
				currentParts = currentParts[2:]
			}
		}
	}

	// parsing successful, so we now know the proper value for .Parts
	info.Parts = currentParts

	// parts look like: resource/resourceName/subresource/other/stuff/we/don't/interpret
	switch {
	case len(info.Parts) >= 3 && !specialVerbsNoSubResources.Has(info.Verb):
		info.Subresource = info.Parts[2]
		fallthrough
	case len(info.Parts) >= 2:
		info.Name = info.Parts[1]
		fallthrough
	case len(info.Parts) >= 1:
		info.Resource = info.Parts[0]
	}

	info.ResourceScope = ri.resolveResourceScope(info)

	if len(info.Name) == 0 && info.Verb == VerbGet {
		info.Verb = VerbList
	}

	if len(info.Name) == 0 && info.Verb == VerbDelete {
		info.Verb = "delete_collection"
	}

	return &info, nil
}

func WithRequestInfo(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, requestInfoKey, info)
}

func InfoFrom(ctx context.Context) (*Info, bool) {
	info, ok := ctx.Value(requestInfoKey).(*Info)
	return info, ok
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

const (
	GlobalScope    = "Global"
	ClusterScope   = "Cluster"
	WorkspaceScope = "Workspace"
	NamespaceScope = "Namespace"
)

func (ri *InfoFactory) resolveResourceScope(info Info) string {
	if info.Namespace != "" {
		return NamespaceScope
	}

	if info.Workspace != "" {
		return WorkspaceScope
	}

	if info.Cluster != "" {
		return ClusterScope
	}

	return GlobalScope
}
