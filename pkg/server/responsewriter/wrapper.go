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
	"net"
	"net/http"
)

// UserProvidedDecorator represents a user (client that uses this package)
// provided decorator that wraps an inner http.ResponseWriter object.
// The user-provided decorator object must return the inner (decorated)
// http.ResponseWriter object via the Unwrap function.
type UserProvidedDecorator interface {
	http.ResponseWriter

	// Unwrap returns the inner http.ResponseWriter object associated
	// with the user-provided decorator
	Unwrap() http.ResponseWriter
}

// WrapForHTTP1Or2 accepts a user-provided decorator for an "inner" http.ResponseWriter
// object and potentially wraps the user-provided decorator with a new http.ResponseWriter
// object that implements http.CloseNotifier, http.Flusher, and/or http.Hijacker by
// delegating to the user-provided decorator (if it implements the relevant method) or
// the inner http.ResponseWriter (otherwise), so that the returned http.ResponseWriter
// object implements the same subset of those interfaces as the inner http.ResponseWriter.
//
// This function handles the following three cases.
//   - The inner http.ResponseWriter implements `http.CloseNotifier`, `http.Flusher`,
//     and `http.Hijacker` (an HTTP/1.1 server provides such a http.ResponseWriter).
//   - The inner http.ResponseWriter implements `http.CloseNotifier` and `http.Flusher`
//     but not `http.Hijacker` (an HTTP/2 server provides such a ResponseWriter).
//   - All the other cases collapsed to this one, in which the given http.ResponseWriter is returned.
//
// There are three applicable terms:
//   - "outer": this is the http.ResponseWriter object returned by the WrapForHTTP1Or2 function.
//   - "user-provided decorator" or "middle": this is the user-provided decorator
//     that decorates an inner http.ResponseWriter object. A user-provided decorator
//     implements the UserProvidedDecorator interface. A user-provided decorator
//     may or may not implement http.CloseNotifier, http.Flusher or http.Hijacker.
//   - "inner": the http.ResponseWriter that the user-provided decorator extends.
func WrapForHTTP1Or2(decorator UserProvidedDecorator) http.ResponseWriter {
	// from go net/http documentation:
	// The default HTTP/1.x and HTTP/2 ResponseWriter implementations support Flusher
	// Handlers should always test for this ability at runtime.
	//
	// The Hijacker interface is implemented by ResponseWriters that allow an HTTP handler
	// to take over the connection.
	// The default ResponseWriter for HTTP/1.x connections supports Hijacker, but HTTP/2 connections
	// intentionally do not. ResponseWriter wrappers may also not support Hijacker.
	// Handlers should always test for this ability at runtime
	//
	// The CloseNotifier interface is implemented by ResponseWriters which allow detecting
	// when the underlying connection has gone away.
	// Deprecated: the CloseNotifier interface predates Go's context package.
	// New code should use Request.Context instead.

	// decorator:
	// type metaResponseWriter struct {
	//   http.ResponseWriter
	//   statusCode int
	//   size       int
	// }
	inner := decorator.Unwrap()
	if innerFlusher, ok := inner.(http.Flusher); ok {
		// for HTTP/2 request, the default http.ResponseWriter object (http2responseWriter)
		// implements http.Flusher
		outerHTTP2 := outerWithFlush{
			UserProvidedDecorator: decorator,
			InnerFlusher:          innerFlusher,
		}

		if innerHijacker, ok := inner.(http.Hijacker); ok {
			return &outerWithFlushAndHijack{
				outerWithFlush: outerHTTP2,
				InnerHijacker:  innerHijacker,
			}
		}
		return outerHTTP2
	}
	return decorator
}

var _ http.Flusher = outerWithFlush{}
var _ http.ResponseWriter = outerWithFlush{}
var _ UserProvidedDecorator = outerWithFlush{}

// outerWithFlush is the outer object that extends the
// user provided decorator with http.Flusher only.
type outerWithFlush struct {
	// UserProvidedDecorator is the user-provided object, it decorates
	// an inner http.ResponseWriter object.
	UserProvidedDecorator

	InnerFlusher http.Flusher
}

func (wr outerWithFlush) Flush() {
	if flusher, ok := wr.UserProvidedDecorator.(http.Flusher); ok {
		flusher.Flush()
		return
	}

	wr.InnerFlusher.Flush()
}

var _ http.Flusher = outerWithFlushAndHijack{}
var _ http.Hijacker = outerWithFlushAndHijack{}
var _ http.ResponseWriter = outerWithFlushAndHijack{}
var _ UserProvidedDecorator = outerWithFlushAndHijack{}

// outerWithCloseNotifyFlushAndHijack is the outer object that extends the
// user-provided decorator with http.Flusher and http.Hijacker.
// This applies to http/1.x requests only.
type outerWithFlushAndHijack struct {
	outerWithFlush

	// http.Hijacker for the inner object
	InnerHijacker http.Hijacker
}

func (wr outerWithFlushAndHijack) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := wr.UserProvidedDecorator.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return wr.InnerHijacker.Hijack()
}
