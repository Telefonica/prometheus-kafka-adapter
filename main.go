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

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/containous/traefik/log"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	log.Info("creating kafka producer")

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   kafkaBrokerList,
		"go.batch.producer":   true,  // Enable batch producer (for increased performance).
		"go.delivery.reports": false, // per-message delivery reports to the Events() channel
	})

	if err != nil {
		logrus.WithError(err).Fatal("couldn't create kafka producer")
	}

	r := gin.New()

	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())

	r.POST("/receive", receiveHandler(p))
	r.GET("/metrics", gin.WrapH(prometheus.UninstrumentedHandler()))

	r.Run()
}
