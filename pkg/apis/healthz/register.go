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

package healthz

import (
	"net/http"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"peta.io/peta/pkg/apis"
)

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := new(restful.WebService)

	ws.Route(ws.GET(DefaultHealthzPath).
		To(h.installHealthz(DefaultHealthzPath)).
		Doc("PETA health check").
		Param(
			ws.QueryParameter("verbose", "Detailed information for out log").
				DataType("string")).
		Operation("health check").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNonResourceAPI}).
		Returns(http.StatusOK, apis.StatusOK, nil))

	ws.Route(ws.GET(DefaultLivezPath).
		To(h.installHealthz(DefaultLivezPath)).
		Doc("PETA liveness check").
		Param(
			ws.QueryParameter("verbose", "Detailed information for out log").
				DataType("string")).
		Operation("liveness check").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNonResourceAPI}).
		Returns(http.StatusOK, apis.StatusOK, nil))

	ws.Route(ws.GET(DefaultReadyzPath).
		To(h.installHealthz(DefaultReadyzPath)).
		Doc("PETA readiness check").
		Param(
			ws.QueryParameter("verbose", "Detailed information for out log").
				DataType("string")).
		Operation("readiness check").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNonResourceAPI}).
		Returns(http.StatusOK, apis.StatusOK, nil))

	container.Add(ws)
	return nil
}
