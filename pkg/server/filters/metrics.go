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
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"peta.io/peta/pkg/server/request"
	"sync"
)

func WithMetrics(next http.Handler) http.Handler {
	var registerOnce sync.Once

	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "peta_request_total",
		Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	}, []string{"verb", "group", "version", "resource", "path"})

	metricsList := []prometheus.Collector{
		requestCount,
	}
	registerOnce.Do(func() {
		registerMetrics(metricsList)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		info, exists := request.InfoFrom(req.Context())
		if exists {
			requestCount.WithLabelValues(
				info.Verb,
				info.APIGroup,
				info.APIVersion,
				info.Resource,
				info.Path).Inc()
		}
		next.ServeHTTP(w, req)
	})
}

func registerMetrics(metricsList []prometheus.Collector) {
	reg := prometheus.DefaultRegisterer
	for _, metric := range metricsList {
		reg.MustRegister(metric)
	}
}
