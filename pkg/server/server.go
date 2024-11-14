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
	"crypto/tls"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/apis"
	healthzhandler "peta.io/peta/pkg/apis/healthz"
	versionhandler "peta.io/peta/pkg/apis/version"
	urlruntime "peta.io/peta/pkg/runtime"
	"peta.io/peta/pkg/server/filters"
	"peta.io/peta/pkg/server/metrics"
	"peta.io/peta/pkg/server/options"
	"peta.io/peta/pkg/server/request"
	"peta.io/peta/pkg/utils/sets"
	"peta.io/peta/pkg/version"
	rt "runtime"
)

// APIServer is PETA server
type APIServer struct {
	Server *http.Server

	*options.APIServerOptions

	container *restful.Container

	VersionInfo *version.Info
}

func NewAPIServer(ctx context.Context, o *options.APIServerOptions) (*APIServer, error) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", o.InsecurePort),
	}

	if o.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(o.TLSCertFile, o.TLSPrivateKey)
		if err != nil {
			return nil, err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		server.Addr = fmt.Sprintf(":%d", o.SecurePort)
	}

	apiServer := &APIServer{
		Server:           server,
		VersionInfo:      version.Get(),
		APIServerOptions: o,
	}

	return apiServer, nil
}

func (s *APIServer) PreRun() error {

	s.container = restful.NewContainer()
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	// install APIs
	s.installPETAAPIs()

	if s.MetricsOptions.Enable {
		s.installMetricsAPIs()
	}

	s.installHealthz()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	combinedHandler, err := s.buildHandlerChain(s.container)
	if err != nil {
		return fmt.Errorf("failed to build handler chain: %w", err)
	}

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
	if s.Server.TLSConfig != nil {
		// TLSConfig not nil, no need to pass certFile & keyFile.
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

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
	requestInfoResolver := &request.InfoFactory{APIPrefixes: sets.New("apis")}

	// TODO: Auditing
	// TODO: Authorization
	// TODO: Authentication
	if s.MetricsOptions.Enable {
		handler = filters.WithMetrics(handler)
	}
	handler = filters.WithRequestInfo(handler, requestInfoResolver)
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

func (s *APIServer) installHealthz() {
	handler := healthzhandler.NewHandler(
		[]healthzhandler.HealthChecker{},
		[]healthzhandler.HealthChecker{},
		[]healthzhandler.HealthChecker{},
	)

	urlruntime.Must(handler.AddToContainer(s.container))
}

func (s *APIServer) installMetricsAPIs() {
	metrics.Install(s.container)
}
