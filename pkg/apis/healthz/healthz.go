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
	"github.com/emicklei/go-restful/v3"
)

const DefaultHealthzPath = "/healthz"

// HealthChecker is named healthz checker.
type HealthChecker interface {
	Name() string
	Check(req *restful.Request) error
}

// PingHealthz returns true automatically when checked
var PingHealthz HealthChecker = ping{}

type ping struct{}

func (p ping) Name() string {
	return "ping"
}

// Check PingHealthz is a healthz check that returns true.
func (p ping) Check(_ *restful.Request) error {
	return nil
}
