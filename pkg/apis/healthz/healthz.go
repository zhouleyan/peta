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
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
	"net/http"
	"peta.io/peta/pkg/apis"
	"strings"
	"sync"
)

const DefaultHealthzPath = "/healthz"
const DefaultLivezPath = "/livez"
const DefaultReadyzPath = "/readyz"

func NewHandler(healthzChecks, livezChecks, readyzChecks []HealthChecker) apis.Handler {
	return &handler{
		healthzChecks: healthzChecks,
		livezChecks:   livezChecks,
		readyzChecks:  readyzChecks,
	}
}

func NewFakeHandler() apis.Handler {
	return &handler{}
}

type handler struct {
	healthzChecks []HealthChecker
	livezChecks   []HealthChecker
	readyzChecks  []HealthChecker
}

func (h *handler) installHealthz(path string) restful.RouteFunction {
	var checks []HealthChecker
	name := strings.Split(strings.TrimPrefix(path, "/"), "/")[0]
	switch name {
	case "healthz":
		checks = h.healthzChecks
	case "livez":
		checks = h.livezChecks
	case "readyz":
		checks = h.readyzChecks
	default:
		checks = []HealthChecker{}
	}
	if len(checks) == 0 {
		klog.V(5).Info("No default health checks specified. Installing the ping handler.")
		checks = []HealthChecker{PingHealthz}
	}
	klog.V(5).Infof("Installing health checkers for (%v): %v", path, formatQuoted(checkerName(checks...)...))

	return handleRootHealth(name, nil, checks...)
}

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

// handleRootHealth returns a restful.RouteFunction that serves the provided checks.
func handleRootHealth(name string, firstTimeHealthy func(), checks ...HealthChecker) restful.RouteFunction {
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

// checkerNames returns the names of the checks in the same order as passed in.
func checkerName(checks ...HealthChecker) []string {
	// accumulate the names of checks for printing them out.
	checkerNames := make([]string, 0, len(checks))
	for _, check := range checks {
		checkerNames = append(checkerNames, check.Name())
	}
	return checkerNames
}

// formatQuoted returns a formatted string of the health check names,
// preserving the order passed in.
func formatQuoted(names ...string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ",")
}
