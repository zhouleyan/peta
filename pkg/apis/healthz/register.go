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
	"bytes"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/apis"
	"strings"
	"sync"
)

func NewHandler(checks ...HealthChecker) apis.Handler {
	return &handler{
		checks: checks,
	}
}

func NewFakeHandler() apis.Handler {
	return &handler{}
}

type handler struct {
	checks []HealthChecker
}

func (h *handler) AddToContainer(container *restful.Container) error {
	if len(h.checks) == 0 {
		klog.V(4).Info("No default health checks specified. Installing the ping handler.")
		h.checks = []HealthChecker{PingHealthz}
	}
	name := strings.Split(strings.TrimPrefix(DefaultHealthzPath, "/"), "/")[0]
	ws := new(restful.WebService)
	ws.Route(ws.GET(DefaultHealthzPath).
		To(handleHealth(name, nil, h.checks...)).
		Doc("PETA health check").
		Param(
			ws.QueryParameter("verbose", "Detailed information for out log").
				DataType("string")).
		Operation("healthcheck").
		Metadata(restfulspec.KeyOpenAPITags, []string{apis.TagNonResourceAPI}).
		Returns(http.StatusOK, apis.StatusOK, nil))

	//container.Handle(DefaultHealthzPath, handleHealth(name, nil, h.checks...))
	container.Add(ws)
	return nil
}

// handleHealth returns a http.HandlerFunc that serves the provided checks.
func handleHealth(name string, firstTimeHealthy func(), checks ...HealthChecker) restful.RouteFunction {
	var notifyOnce sync.Once
	return func(req *restful.Request, response *restful.Response) {
		// failedVerboseLogOutput is for output to the log. It indicates detailed failed output information for the log.
		var failedVerboseLogOutput bytes.Buffer
		var individualCheckOutput bytes.Buffer
		var failedChecks []string
		for _, check := range checks {
			if err := check.Check(req); err != nil {
				// don't include the error since this endpoint is public. If someone wants more detail
				// they should have explicit permission to the detailed checks.
				_, _ = fmt.Fprintf(&individualCheckOutput, "[-]%s failed: reason withheld\n", check.Name())
				// but we do want detailed information for out log
				_, _ = fmt.Fprintf(&failedVerboseLogOutput, "[-]%s failed: %v\n", check.Name(), err)
				failedChecks = append(failedChecks, check.Name())
			} else {
				_, _ = fmt.Fprintf(&individualCheckOutput, "[+]%s ok\n", check.Name())
			}
		}

		// always be verbose on failure
		if len(failedChecks) > 0 {
			klog.V(2).Infof("%s check failed: %s\n%v", strings.Join(failedChecks, ","), name, failedVerboseLogOutput.String())
			http.Error(response, fmt.Sprintf("%s%s check failed", individualCheckOutput.String(), name), http.StatusInternalServerError)
			return
		}

		// signal first time this is healthy
		if firstTimeHealthy != nil {
			notifyOnce.Do(firstTimeHealthy)
		}

		response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		response.Header().Set("X-Content-Type-Options", "nosniff")

		if _, found := req.Request.URL.Query()["verbose"]; !found {
			_, _ = fmt.Fprint(response, "ok")
			return
		}

		_, _ = individualCheckOutput.WriteTo(response)
		_, _ = fmt.Fprintf(response, "%s check passed\n", name)
	}
}
