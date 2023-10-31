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
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"strings"

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

		for topicAndHashKey, metrics := range metricsPerTopic {

			topic, partitionID, err := getPartitionAndTopic(topicAndHashKey)
			if err != nil {
				continue
			}

			part := kafka.TopicPartition{
				Partition: partitionID,
				Topic:     &topic,
			}
			for _, metric := range metrics {
				objectsWritten.Add(float64(1))
				err := producer.Produce(&kafka.Message{
					TopicPartition: part,
					Value:          metric,
				}, nil)

				if err != nil {
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

func getPartitionAndTopic(topic string) (string, int32, error) {
	parts := strings.Split(topic, "|")

	if len(parts) == 1 {
		return parts[0], kafka.PartitionAny, nil
	}
	h := fnv.New32a()
	h.Write([]byte(parts[1]))

	v, ok := topicPartitionCount.Load(parts[0])
	if !ok {
		logrus.WithField("topic", parts[0]).Error("did not find metadata requested topic")
		return topic, kafka.PartitionAny, fmt.Errorf("could not")
	}
	return parts[0], int32(h.Sum32() % uint32(v.(int))), nil
}
