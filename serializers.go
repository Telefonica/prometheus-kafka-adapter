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
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"

	"github.com/linkedin/goavro"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

// Serializer represents an abstract metrics serializer
type Serializer interface {
	Marshal(metric map[string]interface{}, topic string) ([]byte, error)
}

type Metric struct {
	Timestamp string            `json:"timestamp" avro:"timestamp"`
	Value     string            `json:"value" avro:"value"`
	Name      string            `json:"name" avro:"name"`
	Labels    map[string]string `json:"labels" avro:"labels"`
}

// Serialize generates the JSON representation for a given Prometheus metric.
func Serialize(s Serializer, req *prompb.WriteRequest) (map[string][][]byte, error) {
	promBatches.Add(float64(1))
	result := make(map[string][][]byte)

	for _, ts := range req.Timeseries {
		labels := make(map[string]string, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
		}

		t := topic(labels)

		for _, sample := range ts.Samples {
			name := string(labels["__name__"])
			if !filter(name, labels) {
				objectsFiltered.Add(float64(1))
				continue
			}

			epoch := time.Unix(sample.Timestamp/1000, 0).UTC()
			m := map[string]interface{}{
				"timestamp": epoch.Format(time.RFC3339),
				"value":     strconv.FormatFloat(sample.Value, 'f', -1, 64),
				"name":      name,
				"labels":    labels,
			}
			data, err := s.Marshal(m, t)
			if err != nil {
				serializeFailed.Add(float64(1))
				logrus.WithError(err).Errorln("couldn't marshal timeseries")
			}
			serializeTotal.Add(float64(1))
			result[t] = append(result[t], data)
		}
	}

	return result, nil
}

// JSONSerializer represents a metrics serializer that writes JSON
type JSONSerializer struct {
}

func (s *JSONSerializer) Marshal(metric map[string]interface{}, topic string) ([]byte, error) {
	return json.Marshal(metric)
}

func NewJSONSerializer() (*JSONSerializer, error) {
	return &JSONSerializer{}, nil
}

// AvroJSONSerializer represents a metrics serializer that writes Avro-JSON
type AvroJSONSerializer struct {
	codec *goavro.Codec
}

func (s *AvroJSONSerializer) Marshal(metric map[string]interface{}, topic string) ([]byte, error) {
	return s.codec.TextualFromNative(nil, metric)
}

// NewAvroJSONSerializer builds a new instance of the AvroJSONSerializer
func NewAvroJSONSerializer(schemaPath string) (*AvroJSONSerializer, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't read avro schema")
		return nil, err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		logrus.WithError(err).Errorln("couldn't create avro codec")
		return nil, err
	}

	return &AvroJSONSerializer{
		codec: codec,
	}, nil
}

type SchemaRegistrySerializer struct {
	ser *avrov2.Serializer
}

func (s SchemaRegistrySerializer) Marshal(metric map[string]interface{}, topic string) ([]byte, error) {
	m := mapToMetric(metric)
	return s.ser.Serialize(topic, &m)
}

func mapToMetric(metric map[string]interface{}) Metric {
	return Metric{
		Timestamp: metric["timestamp"].(string),
		Value:     metric["value"].(string),
		Name:      metric["name"].(string),
		Labels:    metric["labels"].(map[string]string),
	}
}

func NewAvroSchemaRegistrySerializer(schemaRegistryAPIEndpoint string, schemaRegistryAPIKey string, schemaRegistryAPISecret string) (*SchemaRegistrySerializer, error) {

	client, err := schemaregistry.NewClient(schemaregistry.NewConfigWithBasicAuthentication(
		schemaRegistryAPIEndpoint,
		schemaRegistryAPIKey,
		schemaRegistryAPISecret,
	))

	if err != nil {
		logrus.WithError(err).Errorln("Failed to create schema registry client.")
		return nil, err
	}

	config := avrov2.NewSerializerConfig()
	config.AutoRegisterSchemas = schemaRegistryAutoRegisterSchemas

	ser, err := avrov2.NewSerializer(client, serde.ValueSerde, config)

	if err != nil {
		logrus.WithError(err).Errorln("Failed to create serializer")
		return nil, err
	}

	return &SchemaRegistrySerializer{
		ser: ser,
	}, nil
}

func topic(labels map[string]string) string {
	var buf bytes.Buffer
	if err := topicTemplate.Execute(&buf, labels); err != nil {
		return ""
	}
	return buf.String()
}

func filter(name string, labels map[string]string) bool {
	if len(match) == 0 {
		return true
	}
	mf, ok := match[name]
	if !ok {
		return false
	}

	for _, m := range mf.Metric {
		if len(m.Label) == 0 {
			return true
		}

		labelMatch := true
		for _, label := range m.Label {
			val, ok := labels[label.GetName()]
			if !ok || val != label.GetValue() {
				labelMatch = false
				break
			}
		}

		if labelMatch {
			return true
		}
	}
	return false
}
