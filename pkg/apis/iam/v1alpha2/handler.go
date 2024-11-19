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
	"github.com/emicklei/go-restful/v3"
	"github.com/gofrs/uuid"
	"peta.io/peta/pkg/apis"
	"peta.io/peta/pkg/persistence"
)

type handler struct {
	Storage persistence.Storage
}

type User struct {
	ID   uuid.UUID `db:"id" json:"id"`
	name string
}

func NewHandler(s persistence.Storage) apis.Handler {
	return &handler{Storage: s}
}

func NewFakeHandler() apis.Handler {
	return &handler{}
}

func (h *handler) listUsers(request *restful.Request, response *restful.Response) {
	user := User{}
	err := h.Storage.GetConnection().Find(&user, "1")
	if err != nil {
		apis.HandleInternalError(response, request, err)
		return
	}
	_ = response.WriteAsJson(map[string]string{"v": "ok"})
}
