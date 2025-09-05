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
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/server/request"
	"peta.io/peta/pkg/server/responsewriter"
	"peta.io/peta/pkg/utils/iputils"
)

const (
	_           = iota // ignore first value by assigning to blank identifier
	bKB float64 = 1 << (10 * iota)
	bMB

	Namespace = "peta"
	Subsystem = ""
)

// sizeBuckets is the buckets for request/response size. Here we define a spectrum from 1KB through 1NB up to 10MB.
var sizeBuckets = []float64{1.0 * bKB, 2.0 * bKB, 5.0 * bKB, 10.0 * bKB, 100 * bKB, 500 * bKB, 1.0 * bMB, 2.5 * bMB, 5.0 * bMB, 10.0 * bMB}

func WithMetrics(next http.Handler) http.Handler {
	var registerOnce sync.Once

	labelNames := createLabels()

	requestCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_total",
		Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
	}, labelNames)

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_duration_seconds",
		Help:      "The HTTP request latencies in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, labelNames)

	requestSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_size_bytes",
		Help:      "The HTTP request sizes in bytes",
		Buckets:   sizeBuckets,
	}, labelNames)

	responseSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "response_size_bytes",
		Help:      "The HTTP response sizes in bytes",
		Buckets:   sizeBuckets,
	}, labelNames)

	metricsList := []prometheus.Collector{
		requestCount,
		requestDuration,
		requestSize,
		responseSize,
	}
	registerOnce.Do(func() {
		registerMetrics(metricsList)
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		wrapper := responsewriter.NewMetaResponseWriter(w)
		reqSz := computeApproximateRequestSize(req)
		start := time.Now()

		next.ServeHTTP(responsewriter.WrapForHTTP1Or2(wrapper), req)

		elapsed := time.Since(start)

		info, exists := request.InfoFrom(req.Context())
		if exists {
			values := make([]string, len(labelNames))
			values[0] = info.Path
			values[1] = info.Verb
			values[2] = info.APIGroup
			values[3] = info.APIVersion
			values[4] = info.Resource
			values[5] = req.URL.Path
			values[6] = req.Method
			values[7] = req.Host
			values[8] = strconv.Itoa(wrapper.StatusCode)

			requestCount.WithLabelValues(values...).Inc()
			requestDuration.WithLabelValues(values...).Observe(elapsed.Seconds())
			requestSize.WithLabelValues(values...).Observe(float64(reqSz))
			responseSize.WithLabelValues(values...).Observe(float64(wrapper.Size))
		}

		// Record log for each request
		logWithVerbose := log.Debugf
		// Always log error response
		if wrapper.StatusCode > http.StatusBadRequest {
			logWithVerbose = log.Infof
		}

		logWithVerbose("%s - \"%s %s %s\" %d %d %dms",
			iputils.RemoteIP(req),
			req.Method,
			req.URL,
			req.Proto,
			wrapper.StatusCode,
			wrapper.Size,
			elapsed.Milliseconds(),
		)
	})
}

func registerMetrics(metricsList []prometheus.Collector) {
	reg := prometheus.DefaultRegisterer
	for _, metric := range metricsList {
		reg.MustRegister(metric)
	}
}

func createLabels() []string {
	return []string{"path", "verb", "group", "version", "resource", "url", "method", "host", "code"}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
