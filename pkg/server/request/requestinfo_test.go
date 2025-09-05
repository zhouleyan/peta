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
	"net/http"
	"testing"

	"peta.io/peta/pkg/utils/sets"
)

func newTestRequestInfoResolver() InfoResolver {
	return &InfoFactory{APIPrefixes: sets.New("apis")}
}

func TestRequestInfo(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
	}{
		//{
		//	name:   "version",
		//	url:    "/apis/version.peta.io/v1alpha1/version",
		//	method: http.MethodGet,
		//},
		//{
		//	name:   "namespaces",
		//	url:    "/apis/resource.peta.io/v1alpha1/workspaces/workspace1/namespaces",
		//	method: http.MethodGet,
		//},
		{
			name:   "user",
			url:    "/apis/resource.peta.io/v1alpha1/users/zly/status",
			method: http.MethodGet,
		},
		{
			name:   "user",
			url:    "/clusters/c1/apis/resource.peta.io/v1alpha1/users/zly/status",
			method: http.MethodGet,
		},
		{
			name:   "user",
			url:    "/clusters/c1/apis/resource.peta.io/v1alpha1/workspaces/w1/users/zly/status/foo/bar/baz",
			method: http.MethodGet,
		},
		{
			name:   "user",
			url:    "/clusters/c1/apis/resource.peta.io/v1alpha1/workspaces/w1/namespaces/n1/users/zly/status",
			method: http.MethodGet,
		},
	}

	requestInfoResolver := newTestRequestInfoResolver()

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			req, err := http.NewRequest(test.method, test.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			info, err := requestInfoResolver.NewRequestInfo(req)
			if err != nil {
				t.Errorf("%s", err)
			}
			t.Logf("%+v", info)
		})
	}
}
