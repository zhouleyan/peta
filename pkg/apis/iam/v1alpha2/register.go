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

package v1alpha2

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"peta.io/peta/pkg/apis"
)

const (
	GroupName = "iam.peta.io"
)

var GroupVersion = apis.GroupVersion{
	Group:   GroupName,
	Version: "v1alpha2",
}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := apis.NewWebService(GroupVersion)

	ws.Route(ws.GET("/users").
		Doc("list users").
		Operation("users-list").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNamespacedResources}).
		Notes("list PETA users").
		To(h.listUsers))

	container.Add(ws)
	return nil
}
