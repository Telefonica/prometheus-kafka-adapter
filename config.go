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
	"os"
	"strconv"
	"strings"
	"text/template"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"
)

var (
	kafkaBrokerList                   = "kafka:9092"
	kafkaTopic                        = "metrics"
	topicTemplate                     *template.Template
	match                             = make(map[string]*dto.MetricFamily, 0)
	basicauth                         = false
	basicauthUsername                 = ""
	basicauthPassword                 = ""
	kafkaCompression                  = "none"
	kafkaBatchNumMessages             = "10000"
	kafkaBatchSize                    = "1000000"
	kafkaLingerMs                     = "5"
	kafkaSslClientCertFile            = ""
	kafkaSslClientKeyFile             = ""
	kafkaSslClientKeyPass             = ""
	kafkaSslCACertFile                = ""
	kafkaSecurityProtocol             = ""
	kafkaSaslMechanism                = ""
	kafkaSaslUsername                 = ""
	kafkaSaslPassword                 = ""
	serializer                        Serializer
	kafkaAcks                         = "all"
	schemaRegistryUrl                 = ""
	schemaRegistryUsername            = ""
	schemaRegistryPassword            = ""
	schemaRegistryAutoRegisterSchemas = false
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if value := os.Getenv("LOG_LEVEL"); value != "" {
		logrus.SetLevel(parseLogLevel(value))
	}

	if value := os.Getenv("KAFKA_BROKER_LIST"); value != "" {
		kafkaBrokerList = value
	}

	if value := os.Getenv("KAFKA_TOPIC"); value != "" {
		kafkaTopic = value
	}

	if value := os.Getenv("BASIC_AUTH_USERNAME"); value != "" {
		basicauth = true
		basicauthUsername = value
	}

	if value := os.Getenv("BASIC_AUTH_PASSWORD"); value != "" {
		basicauthPassword = value
	}

	if value := os.Getenv("KAFKA_COMPRESSION"); value != "" {
		kafkaCompression = value
	}

	if value := os.Getenv("KAFKA_BATCH_NUM_MESSAGES"); value != "" {
		kafkaBatchNumMessages = value
	}

	if value := os.Getenv("KAFKA_BATCH_SIZE"); value != "" {
		kafkaBatchSize = value
	}

	if value := os.Getenv("KAFKA_LINGER_MS"); value != "" {
		kafkaLingerMs = value
	}

	if value := os.Getenv("KAFKA_SSL_CLIENT_CERT_FILE"); value != "" {
		kafkaSslClientCertFile = value
	}

	if value := os.Getenv("KAFKA_SSL_CLIENT_KEY_FILE"); value != "" {
		kafkaSslClientKeyFile = value
	}

	if value := os.Getenv("KAFKA_SSL_CLIENT_KEY_PASS"); value != "" {
		kafkaSslClientKeyPass = value
	}

	if value := os.Getenv("KAFKA_SSL_CA_CERT_FILE"); value != "" {
		kafkaSslCACertFile = value
	}

	if value := os.Getenv("KAFKA_SECURITY_PROTOCOL"); value != "" {
		kafkaSecurityProtocol = value
	}

	if value := os.Getenv("KAFKA_SASL_MECHANISM"); value != "" {
		kafkaSaslMechanism = value
	}

	if value := os.Getenv("KAFKA_SASL_USERNAME"); value != "" {
		kafkaSaslUsername = value
	}

	if value := os.Getenv("KAFKA_SASL_PASSWORD"); value != "" {
		kafkaSaslPassword = value
	}
	if value := os.Getenv("KAFKA_ACKS"); value != "" {
		kafkaAcks = value
	}

	if value := os.Getenv("MATCH"); value != "" {
		matchList, err := parseMatchList(value)
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't parse the match rules")
		}
		match = matchList
	}

	if value := os.Getenv("SCHEMA_REGISTRY_URL"); value != "" {
		schemaRegistryUrl = value
	}
	if value := os.Getenv("SCHEMA_REGISTRY_USERNAME"); value != "" {
		schemaRegistryUsername = value
	}
	if value := os.Getenv("SCHEMA_REGISTRY_PASSWORD"); value != "" {
		schemaRegistryPassword = value
	}
	if value := os.Getenv("SCHEMA_REGISTRY_AUTO_REGISTRY_SCHEMAS"); value != "" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			logrus.WithError(err).Fatalln("couldn't parse SCHEMA_REGISTRY_AUTO_REGISTRY_SCHEMAS to bool, using false")
			v = false
		}
		schemaRegistryAutoRegisterSchemas = v
	}

	var err error
	serializer, err = parseSerializationFormat(os.Getenv("SERIALIZATION_FORMAT"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't create a metrics serializer")
	}

	topicTemplate, err = parseTopicTemplate(kafkaTopic)
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't parse the topic template")
	}
}

func parseMatchList(text string) (map[string]*dto.MetricFamily, error) {
	var matchRules []string
	err := yaml.Unmarshal([]byte(text), &matchRules)
	if err != nil {
		return nil, err
	}
	var metricsList []string
	for _, v := range matchRules {
		metricsList = append(metricsList, fmt.Sprintf("%s 0\n", v))
	}

	metricsText := strings.Join(metricsList, "")

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(metricsText))
	if err != nil {
		return nil, fmt.Errorf("couldn't parse match rules: %s", err)
	}
	return metricFamilies, nil
}

func parseLogLevel(value string) logrus.Level {
	level, err := logrus.ParseLevel(value)

	if err != nil {
		logrus.WithField("log-level-value", value).Warningln("invalid log level from env var, using info")
		return logrus.InfoLevel
	}

	return level
}

func parseSerializationFormat(value string) (Serializer, error) {
	switch value {
	case "json":
		return NewJSONSerializer()
	case "avro-json":
		return NewAvroJSONSerializer("schemas/metric.avsc")
	case "avro-schema-registry":
		return NewAvroSchemaRegistrySerializer(schemaRegistryUrl, schemaRegistryUsername, schemaRegistryPassword)
	default:
		logrus.WithField("serialization-format-value", value).Warningln("invalid serialization format, using json")
		return NewJSONSerializer()
	}
}

func parseTopicTemplate(tpl string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"replace": func(old, new, src string) string {
			return strings.Replace(src, old, new, -1)
		},
		"substring": func(start, end int, s string) string {
			if start < 0 {
				start = 0
			}
			if end < 0 || end > len(s) {
				end = len(s)
			}
			if start >= end {
				panic("template function - substring: start is bigger (or equal) than end. That will produce an empty string.")
			}
			return s[start:end]
		},
	}
	return template.New("topic").Funcs(funcMap).Parse(tpl)
}
