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

package version

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"peta.io/peta/pkg/apis"
	"peta.io/peta/pkg/version"
)

var GroupVersion = apis.GroupVersion{
	Group:   "version.peta.io",
	Version: "",
}

func NewHandler(versionInfo *version.Info) apis.Handler {
	return &handler{versionInfo: versionInfo}
}

func NewFakeHandler() apis.Handler {
	return &handler{}
}

type handler struct {
	versionInfo *version.Info
}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := apis.NewWebService(GroupVersion)

	versionFunc := func(request *restful.Request, response *restful.Response) {
		v := version.Get()
		_ = response.WriteAsJson(v)
	}

	ws.Route(ws.GET("/version").
		To(versionFunc).
		Doc("PETA version info").
		Operation("version").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNonResourceAPI}).
		Returns(http.StatusOK, apis.StatusOK, version.Info{}))

	container.Add(ws)

	return nil
}
