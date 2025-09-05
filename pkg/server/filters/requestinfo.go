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

package filters

import (
	"fmt"
	"net/http"

	"peta.io/peta/pkg/apis"
	"peta.io/peta/pkg/server/request"
)

func WithRequestInfo(next http.Handler, resolver request.InfoResolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		info, err := resolver.NewRequestInfo(req)
		if err != nil {
			apis.InternalError(w, req, fmt.Errorf("failed to crate request info: %v", err))
		}

		*req = *req.WithContext(request.WithRequestInfo(ctx, info))
		next.ServeHTTP(w, req)
	})
}
