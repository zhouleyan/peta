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

package responsewriter

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

var _ http.ResponseWriter = &MetaResponseWriter{}
var _ UserProvidedDecorator = &MetaResponseWriter{}

type MetaResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

func (r *MetaResponseWriter) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func NewMetaResponseWriter(w http.ResponseWriter) *MetaResponseWriter {
	return &MetaResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (r *MetaResponseWriter) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *MetaResponseWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *MetaResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	if err != nil {
		return size, err
	}
	return size, nil
}

func (r *MetaResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter doesn't support Hijacker interface")
	}
	return hijacker.Hijack()
}
