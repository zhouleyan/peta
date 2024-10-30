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

package apis

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"strings"
)

const (
	APIRootPath = "/apis"
)

// Container holds all webservice of api-server
var Container = restful.NewContainer()

type Handler interface {
	AddToContainer(c *restful.Container) error
}

// GroupVersion contains the "group" and the "version", which uniquely identifies the API.
type GroupVersion struct {
	Group   string
	Version string
}

const MimeMergePatchJson = "application/merge-patch+json"
const MimeJsonPatchJson = "application/json-patch+json"
const MimeMultipartFormData = "multipart/form-data"

func init() {
	restful.RegisterEntityAccessor(MimeMergePatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
	restful.RegisterEntityAccessor(MimeJsonPatchJson, restful.NewEntityAccessorJSON(restful.MIME_JSON))
}

func NewWebService(gv GroupVersion) *restful.WebService {
	webservice := new(restful.WebService)
	// the GroupVersion might be empty, we need to remove the final /
	webservice.Path(strings.TrimRight(APIRootPath+"/"+gv.String(), "/")).
		Produces(restful.MIME_JSON)
	return webservice
}

// Empty returns true of group and version are empty
func (gv GroupVersion) Empty() bool { return len(gv.Group) == 0 && len(gv.Version) == 0 }

// String puts "group" and "version" into a single "group/version" string.
func (gv GroupVersion) String() string {
	if len(gv.Group) > 0 {
		return fmt.Sprintf("%s/%s", gv.Group, gv.Version)
	}
	return gv.Version
}
