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

package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/apis"
	versionhandler "peta.io/peta/pkg/apis/version"
	"peta.io/peta/pkg/config"
	urlruntime "peta.io/peta/pkg/runtime"
	"peta.io/peta/pkg/version"
	rt "runtime"
)

// APIServer is PETA server
type APIServer struct {
	Server *http.Server

	container *restful.Container

	config *config.Config

	VersionInfo *version.Info
}

func NewAPIServer(ctx context.Context) (*APIServer, error) {
	apiServer := &APIServer{
		VersionInfo: version.Get(),
	}

	return apiServer, nil
}

func (s *APIServer) PreRun() error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 9090),
	}
	s.Server = server

	s.container = restful.NewContainer()
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	// install APIs
	s.installPETAAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	combinedHandler, err := s.buildHandlerChain(s.container)
	if err != nil {
		return fmt.Errorf("failed to build handler chain: %w", err)
	}

	//s.Server.Handler = filters.WithGlobalFilter(combinedHandler)
	s.Server.Handler = combinedHandler

	return nil
}

func (s *APIServer) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		klog.V(0).Info("Server shutting down")
		if err := s.Server.Shutdown(ctx); err != nil {
			klog.Errorf("failed to shutdown server: %s", err)
		}
	}()

	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	err = s.Server.ListenAndServe()

	return err
}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (s *APIServer) buildHandlerChain(handler http.Handler) (http.Handler, error) {
	return handler, nil
}

func (s *APIServer) installPETAAPIs() {
	handlers := []apis.Handler{
		versionhandler.NewHandler(s.VersionInfo),
	}

	for _, handler := range handlers {
		urlruntime.Must(handler.AddToContainer(s.container))
	}
}
