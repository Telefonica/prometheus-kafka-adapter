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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/snappy"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/prometheus/prompb"
)

func receiveHandler(producer *kafka.Producer, serializer Serializer) func(c *gin.Context) {
	return func(c *gin.Context) {

		httpRequestsTotal.Add(float64(1))

		compressed, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithError(err).Error("couldn't read body")
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			logrus.WithError(err).Error("couldn't decompress body")
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			logrus.WithError(err).Error("couldn't unmarshal body")
			return
		}

		metricsPerTopic, err := processWriteRequest(&req)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithError(err).Error("couldn't process write request")
			return
		}

		for topic, metrics := range metricsPerTopic {
			t := topic
			part := kafka.TopicPartition{
				Partition: kafka.PartitionAny,
				Topic:     &t,
			}
			for _, metric := range metrics {
				objectsWritten.Add(float64(1))
				err := producer.Produce(&kafka.Message{
					TopicPartition: part,
					Value:          metric,
				}, nil)

				go func() {
					for event := range producer.Events() {
						switch ev := event.(type) {
						case *kafka.Message:
							message := ev
							if message.TopicPartition.Error != nil {
								logrus.WithError(message.TopicPartition.Error).Errorf("failed to deliver message: %v",
									message.TopicPartition)
							} else {
								logrus.Debugf("delivered to topic %s [%d] at offset %v",
									*message.TopicPartition.Topic,
									message.TopicPartition.Partition,
									message.TopicPartition.Offset)
							}
						case kafka.Error:
							logrus.WithError(err).Errorf("Error: %v", ev)
						default:
							logrus.Infof("Ignored event: %s", ev)
						}
					}
				}()

				if err != nil {
					if err.(kafka.Error).Code() == kafka.ErrQueueFull {
						// Producer queue is full, wait 1s for messages to delivered
						// Maybe we should fail fast? As we are losing data...
						logrus.Warning("producer queue is full, waiting 1s")
						time.Sleep(time.Second)
					}

					objectsFailed.Add(float64(1))
					c.AbortWithStatus(http.StatusInternalServerError)
					logrus.WithError(err).Debug(fmt.Sprintf("Failing metric %v", metric))
					logrus.WithError(err).Error(fmt.Sprintf("couldn't produce message in kafka topic %v", topic))
					return
				}
			}
		}

	}
}
