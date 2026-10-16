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
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

// fakeProducer implements the kafkaProducer interface for testing. It records
// how many times Produce was called and returns a preconfigured error.
type fakeProducer struct {
	err   error
	calls int
}

func (f *fakeProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	f.calls++
	return f.err
}

// newReceiveRequest builds a snappy-compressed, protobuf-encoded remote_write
// request body containing a single sample, matching what Prometheus sends to
// the /receive endpoint.
func newReceiveRequest(t *testing.T) *http.Request {
	t.Helper()

	wr := &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "up"},
					{Name: "job", Value: "test"},
				},
				Samples: []prompb.Sample{
					{Timestamp: 0, Value: 1},
				},
			},
		},
	}

	data, err := proto.Marshal(wr)
	assert.Nil(t, err)

	compressed := snappy.Encode(nil, data)
	req := httptest.NewRequest(http.MethodPost, "/receive", bytes.NewReader(compressed))
	return req
}

// setupHandlerTest initialises the package-level globals that processWriteRequest
// relies on (serializer + topic template) so the handler produces exactly one
// message to a known topic.
func setupHandlerTest(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var err error
	serializer, err = NewJSONSerializer()
	assert.Nil(t, err)

	topicTemplate, err = parseTopicTemplate("metrics")
	assert.Nil(t, err)

	// No filter rules are configured by default (match is an empty map at the
	// package level), so the sample is not dropped by filter().
}

func runReceive(t *testing.T, producer kafkaProducer) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newReceiveRequest(t)

	receiveHandler(producer, serializer)(c)
	return w
}

func TestReceiveHandlerSkipsUnknownTopic(t *testing.T) {
	setupHandlerTest(t)

	producer := &fakeProducer{
		err: kafka.NewError(kafka.ErrUnknownTopic, "unknown topic", false),
	}
	w := runReceive(t, producer)

	assert.Equal(t, 1, producer.calls, "producer should have been called once")
	assert.NotEqual(t, http.StatusInternalServerError, w.Code,
		"an unknown topic should be skipped, not abort the batch with 500")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiveHandlerSkipsUnknownTopicOrPart(t *testing.T) {
	setupHandlerTest(t)

	producer := &fakeProducer{
		err: kafka.NewError(kafka.ErrUnknownTopicOrPart, "unknown topic or partition", false),
	}
	w := runReceive(t, producer)

	assert.Equal(t, 1, producer.calls, "producer should have been called once")
	assert.NotEqual(t, http.StatusInternalServerError, w.Code,
		"an unknown topic-or-partition should be skipped, not abort the batch with 500")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReceiveHandlerAbortsOnGenericError(t *testing.T) {
	setupHandlerTest(t)

	producer := &fakeProducer{
		err: kafka.NewError(kafka.ErrAllBrokersDown, "all brokers down", false),
	}
	w := runReceive(t, producer)

	assert.Equal(t, 1, producer.calls, "producer should have been called once")
	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"a generic produce error must still abort the request with 500")
}

func TestReceiveHandlerSucceedsWhenProduceOK(t *testing.T) {
	setupHandlerTest(t)

	producer := &fakeProducer{err: nil}
	w := runReceive(t, producer)

	assert.Equal(t, 1, producer.calls, "producer should have been called once")
	assert.Equal(t, http.StatusOK, w.Code)
}
