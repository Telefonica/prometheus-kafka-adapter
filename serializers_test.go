package main

import (
	"math"
	"testing"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

func NewWriteRequest() *prompb.WriteRequest {
	return &prompb.WriteRequest{
		Timeseries: []*prompb.TimeSeries{
			&prompb.TimeSeries{
				Labels: []*prompb.Label{
					&prompb.Label{Name: "__name__", Value: "foo"},
					&prompb.Label{Name: "labelfoo", Value: "label-bar"},
				},
				Samples: []*prompb.Sample{
					&prompb.Sample{Timestamp: 0, Value: 456},
					&prompb.Sample{Timestamp: 10000, Value: math.Inf(1)},
				},
			},
		},
	}
}

func TestSerializeEmptyTimeseriesToJSON(t *testing.T) {
	request := &prompb.WriteRequest{}
	serializer, err := NewJSONSerializer()
	assert.Nil(t, err)

	data, err := Serialize(serializer, request)
	assert.Nil(t, err)
	assert.Len(t, data, 0)
	assert.NotNil(t, data)
}

func TestSerializeToJSON(t *testing.T) {
	serializer, err := NewJSONSerializer()
	assert.Nil(t, err)

	writeRequest := NewWriteRequest()
	output, err := Serialize(serializer, writeRequest)
	assert.Len(t, output["metrics"], 2)
	assert.Nil(t, err)

	expectedSamples := []string{
		"{\"value\":\"456\",\"timestamp\":\"1970-01-01T00:00:00Z\",\"name\":\"foo\",\"labels\":{\"__name__\":\"foo\",\"labelfoo\":\"label-bar\"}}",
		"{\"value\":\"+Inf\",\"timestamp\":\"1970-01-01T00:00:10Z\",\"name\":\"foo\",\"labels\":{\"__name__\":\"foo\",\"labelfoo\":\"label-bar\"}}",
	}

	for i, metric := range output["metrics"] {
		assert.JSONEqf(t, expectedSamples[i], string(metric[:]), "wrong json serialization found")
	}
}

func TestSerializeEmptyTimeseriesToAvroJSON(t *testing.T) {
	request := &prompb.WriteRequest{}
	serializer, err := NewAvroJSONSerializer("schemas/metric.avsc")
	assert.Nil(t, err)

	data, err := Serialize(serializer, request)
	assert.Nil(t, err)
	assert.Len(t, data, 0)
	assert.NotNil(t, data)
}

func TestSerializeToAvro(t *testing.T) {
	serializer, err := NewAvroJSONSerializer("schemas/metric.avsc")
	assert.Nil(t, err)

	writeRequest := NewWriteRequest()
	output, err := Serialize(serializer, writeRequest)
	assert.Len(t, output["metrics"], 2)
	assert.Nil(t, err)

	expectedSamples := []string{
		"{\"value\":\"456\",\"timestamp\":\"1970-01-01T00:00:00Z\",\"name\":\"foo\",\"labels\":{\"__name__\":\"foo\",\"labelfoo\":\"label-bar\"}}",
		"{\"value\":\"+Inf\",\"timestamp\":\"1970-01-01T00:00:10Z\",\"name\":\"foo\",\"labels\":{\"__name__\":\"foo\",\"labelfoo\":\"label-bar\"}}",
	}

	for i, metric := range output["metrics"] {
		assert.JSONEqf(t, expectedSamples[i], string(metric[:]), "wrong json serialization found")
	}
}

func TestTemplatedTopic(t *testing.T) {
	var err error
	topicTemplate, err = parseTopicTemplate("{{ index . \"labelfoo\" | replace \"bar\" \"foo\" | substring 6 -1 }}")
	assert.Nil(t, err)
	serializer, err := NewJSONSerializer()
	assert.Nil(t, err)

	writeRequest := NewWriteRequest()
	output, err := Serialize(serializer, writeRequest)

	for k, _ := range output {
		assert.Equal(t, "foo", k, "templated topic failed")
	}
}

func BenchmarkSerializeToAvroJSON(b *testing.B) {
	serializer, _ := NewAvroJSONSerializer("schemas/metric.avsc")
	writeRequest := NewWriteRequest()

	for n := 0; n < 20000; n++ {
		Serialize(serializer, writeRequest)
	}
}

func BenchmarkSerializeToJSON(b *testing.B) {
	serializer, _ := NewJSONSerializer()
	writeRequest := NewWriteRequest()

	for n := 0; n < 20000; n++ {
		Serialize(serializer, writeRequest)
	}
}
