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
	"encoding/json"
	"strconv"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"
)

func processWriteRequest(req *prompb.WriteRequest) ([][]byte, error) {
	logrus.WithField("var", req).Debugln()
	result := [][]byte{}

	for _, ts := range req.Timeseries {
		labels := make(model.Metric, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}

		for _, sample := range ts.Samples {
			metric := make(map[string]interface{}, len(labels)+2)
			metric["__value__"] = strconv.FormatFloat(sample.Value, 'f', -1, 64)
			metric["__timestamp__"] = sample.Timestamp

			for key, value := range labels {
				metric[string(key)] = value
			}

			data, err := json.Marshal(metric)
			if err != nil {
				logrus.WithError(err).Errorln("couldn't proccess timeseries")
			}

			result = append(result, data)
		}
	}

	return result, nil
}
