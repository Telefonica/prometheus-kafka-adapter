// Copyright 2018 Telefónica
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "github.com/prometheus/client_golang/prometheus"

var (
	httpRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all http requests",
		})
	promBatches = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "incoming_prometheus_batches_total",
			Help: "Count of incoming prometheus batches (to be broken into individual metrics)",
		})
	serializeTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "serialized_total",
			Help: "Count of all serialization requests",
		})
	serializeFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "serialized_failed_total",
			Help: "Count of all serialization failures",
		})
	objectsFiltered = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_filtered_total",
			Help: "Count of all filter attempts",
		})
	objectsWritten = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_written_total",
			Help: "Count of all objects written to Kafka",
		})
	objectsFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "objects_failed_total",
			Help: "Count of all objects write failures to Kafka",
		})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(promBatches)
	prometheus.MustRegister(serializeTotal)
	prometheus.MustRegister(serializeFailed)
	prometheus.MustRegister(objectsFiltered)
	prometheus.MustRegister(objectsFailed)
	prometheus.MustRegister(objectsWritten)
}
