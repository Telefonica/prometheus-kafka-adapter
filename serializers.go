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
	"io/ioutil"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"

	"github.com/linkedin/goavro"
)

// Serializer represents an abstract metrics serializer
type Serializer interface {
	Marshal(metric map[string]interface{}) ([]byte, error)
}

// Serialize generates the JSON representation for a given Prometheus metric.
func Serialize(s Serializer, req *prompb.WriteRequest) (map[string][][]byte, error) {
	result := make(map[string][][]byte)

	for _, ts := range req.Timeseries {
		labels := make(map[string]string, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
		}

		t := topic(labels)

		for _, sample := range ts.Samples {
			epoch := time.Unix(sample.Timestamp/1000, 0).UTC()

			m := map[string]interface{}{
				"timestamp": epoch.Format(time.RFC3339),
				"value":     strconv.FormatFloat(sample.Value, 'f', -1, 64),
				"name":      string(labels["__name__"]),
				"labels":    labels,
			}

			data, err := s.Marshal(m)
			if err != nil {
				logrus.WithError(err).Errorln("couldn't marshal timeseries")
			}

			result[t] = append(result[t], data)
		}
	}

	return result, nil
}

// JSONSerializer represents a metrics serializer that writes JSON
type JSONSerializer struct {
}

func (s *JSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
	return json.Marshal(metric)
}

func NewJSONSerializer() (*JSONSerializer, error) {
	return &JSONSerializer{}, nil
}

// AvroJSONSerializer represents a metrics serializer that writes Avro-JSON
type AvroJSONSerializer struct {
	codec *goavro.Codec
}

func (s *AvroJSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
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

func topic(labels map[string]string) string {
	var buf bytes.Buffer
	if err := topicTemplate.Execute(&buf, labels); err != nil {
		return ""
	}
	return buf.String()
}
