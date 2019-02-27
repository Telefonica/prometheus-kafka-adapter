# prometheus-kafka-adapter

[![Build Status](https://travis-ci.org/Telefonica/prometheus-kafka-adapter.svg?branch=master)](https://travis-ci.org/Telefonica/prometheus-kafka-adapter)

Prometheus-kafka-adapter is a service which receives [Prometheus](https://github.com/prometheus) metrics through [`remote_write`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), marshal into JSON and sends them into [Kafka](https://github.com/apache/kafka).

## motivation

We use `prometheus-kafka-adapter` internally at Telefonica for dumping Prometheus metrics into an object storage in diferent clouds, through [Kafka](https://github.com/apache/kafka) and Kafka-Connect.

## output

It is able to write JSON or Avro-JSON messages in a kafka topic, depending on the `SERIALIZATION_FORMAT` configuration variable.

### JSON

```json
{
  "timestamp": "1970-01-01T00:00:00Z",
  "value": "9876543210",
  "name": "up",

  "labels": {
    "__name__": "up",
    "label1": "value1",
    "label2": "value2"
  }
}
```

`timestamp` and `value` are reserved values, and can't be used as label names. `__name__` is a special label that defines the name of the metric and is copied as `name` to the top level for convenience.

### Avro JSON

The Avro-JSON serialization is the same. See the [Avro schema](./schemas/metric.avsc).

## configuration

### prometheus-kafka-adapter

There is a docker image `telefonica/prometheus-kafka-adapter:1.2.0` [available on Docker Hub](https://hub.docker.com/r/telefonica/prometheus-kafka-adapter/).

Prometheus-kafka-adapter listens for metrics coming from Prometheus and sends them to Kafka. This behaviour can be configured with the following environment variables:

- `KAFKA_BROKER_LIST`: defines kafka endpoint and port, defaults to `kafka:9092`.
- `KAFKA_TOPIC`: defines kafka topic to be used, defaults to `metrics`.
- `KAFKA_COMPRESSION`: defines the compression type to be used, defaults to `none`.
- `KAFKA_BATCH_NUM_MESSAGES`: defines the number of messages to batch write, defaults to `10000`.
- `SERIALIZATION_FORMAT`: defines the serialization format, can be `json`, `avro-json`, defaults to `json`.
- `PORT`: defines http port to listen, defaults to `8080`, used directly by [gin](https://github.com/gin-gonic/gin).
- `LOG_LEVEL`: defines log level for [`logrus`](https://github.com/sirupsen/logrus), can be `debug`, `info`, `warn`, `error`, `fatal` or `panic`, defaults to `info`.
- `GIN_MODE`: manage [gin](https://github.com/gin-gonic/gin) debug logging, can be `debug` or `release`.

### prometheus

Prometheus needs to have a `remote_write` url configured, pointing to the '/receive' endpoint of the host and port where the prometheus-kafka-adapter service is running. For example:

```yaml
remote_write:
  - url: "http://prometheus-kafka-adapter:8080/receive"
```

## development

```
go test
go build
```

## contributing

With issues:
  - Use the search tool before opening a new issue.
  - Please provide source code and commit sha if you found a bug.
  - Review existing issues and provide feedback or react to them.

With pull requests:
  - Open your pull request against `master`
  - It should pass all tests in the continuous integrations system (TravisCI).
  - You should add/modify tests to cover your proposed code changes.
  - If your pull request contains a new feature, please document it on the README.


## license

Copyright 2018 [Telefónica](https://www.telefonica.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
