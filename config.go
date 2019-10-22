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
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var (
	kafkaBrokerList   = "kafka:9092"
	kafkaTopic        = "metrics"
	basicauth         = false
	basicauthUsername = ""
	basicauthPassword = ""
	kafkaPartition    = kafka.TopicPartition{
		Topic:     &kafkaTopic,
		Partition: kafka.PartitionAny,
	}
	kafkaCompression       = "none"
	kafkaBatchNumMessages  = "10000"
	kafkaSslClientCertFile = ""
	kafkaSslClientKeyFile  = ""
	kafkaSslClientKeyPass  = ""
	kafkaSslCACertFile     = ""
	kafkaSslValidation     = true
	serializer             Serializer
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if value := os.Getenv("LOG_LEVEL"); value == "" {
		logrus.SetLevel(parseLogLevel(value))
	}

	if value := os.Getenv("KAFKA_BROKER_LIST"); value != "" {
		kafkaBrokerList = value
	}

	if value := os.Getenv("KAFKA_TOPIC"); value != "" {
		kafkaTopic = value

		kafkaPartition = kafka.TopicPartition{
			Topic:     &kafkaTopic,
			Partition: kafka.PartitionAny,
		}
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

	var err error
	serializer, err = parseSerializationFormat(os.Getenv("SERIALIZATION_FORMAT"))
	if err != nil {
		logrus.WithError(err).Fatalln("couldn't create a metrics serializer")
	}
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
	default:
		logrus.WithField("serialization-format-value", value).Warningln("invalid serialization format, using json")
		return NewJSONSerializer()
	}
}
