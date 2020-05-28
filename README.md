# prometheus-kafka-adapter

[![Build Status](https://travis-ci.org/Telefonica/prometheus-kafka-adapter.svg?branch=master)](https://travis-ci.org/Telefonica/prometheus-kafka-adapter)

Prometheus-kafka-adapter is a service which receives [Prometheus](https://github.com/prometheus) metrics through [`remote_write`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), marshal into JSON and sends them into [Kafka](https://github.com/apache/kafka).

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

There is a docker image `telefonica/prometheus-kafka-adapter:1.6.0` [available on Docker Hub](https://hub.docker.com/r/telefonica/prometheus-kafka-adapter/).

Prometheus-kafka-adapter listens for metrics coming from Prometheus and sends them to Kafka. This behaviour can be configured with the following environment variables:

- `KAFKA_BROKER_LIST`: defines kafka endpoint and port, defaults to `kafka:9092`.
- `KAFKA_TOPIC`: defines kafka topic to be used, defaults to `metrics`. Could use go template, labels are passed (as a map) to the template: e.g: `metrics.{{ index . "__name__" }}` to use per-metric topic. Two template functions are available: replace (`{{ index . "__name__" | replace "message" "msg" }}`) and substring (`{{ index . "__name__" | substring 0 5 }}`)
- `KAFKA_COMPRESSION`: defines the compression type to be used, defaults to `none`.
- `KAFKA_BATCH_NUM_MESSAGES`: defines the number of messages to batch write, defaults to `10000`.
- `SERIALIZATION_FORMAT`: defines the serialization format, can be `json`, `avro-json`, defaults to `json`.
- `PORT`: defines http port to listen, defaults to `8080`, used directly by [gin](https://github.com/gin-gonic/gin).
- `BASIC_AUTH_USERNAME`: basic auth username to be used for receive endpoint, defaults is no basic auth.
- `BASIC_AUTH_PASSWORD`: basic auth password to be used for receive endpoint, defaults is no basic auth.
- `LOG_LEVEL`: defines log level for [`logrus`](https://github.com/sirupsen/logrus), can be `debug`, `info`, `warn`, `error`, `fatal` or `panic`, defaults to `info`.
- `GIN_MODE`: manage [gin](https://github.com/gin-gonic/gin) debug logging, can be `debug` or `release`.

To connect to Kafka over SSL define the following additonal environment variables:

- `KAFKA_SSL_CLIENT_CERT_FILE`: Kafka SSL client certificate file, defaults to `""`
- `KAFKA_SSL_CLIENT_KEY_FILE`: Kafka SSL client certificate key file, defaults to `""`
- `KAFKA_SSL_CLIENT_KEY_PASS`: Kafka SSL client certificate key password (optional), defaults to `""`
- `KAFKA_SSL_CA_CERT_FILE`: Kafka SSL broker CA certificate file, defaults to `""`

When deployed in a Kubernetes cluster using Helm and using a Kafka external to the cluster, it might be necessary to define the kafka hostname resolution locally (this fills the /etc/hosts of the container). Use a custom values.yaml file with section `hostAliases` (as mentioned in default values.yaml).

### prometheus

Prometheus needs to have a `remote_write` url configured, pointing to the '/receive' endpoint of the host and port where the prometheus-kafka-adapter service is running. For example:

```yaml
remote_write:
  - url: "http://prometheus-kafka-adapter:8080/receive"
```

When deployed in a Kubernetes cluster using Helm and using an external Prometheus, it might be necessary to expose prometheus-kafka-adapter input port as a node port. Use a custom values.yaml file to set `service.type: NodePort` and `service.nodeport: <PortNumber>` (see comments in default values.yaml)

## development

```
go test
go build
```

## contributing

With issues:
  - Use the search tool before opening a new issue.
  - Please provide source code and commit sha if you find a bug.
  - Review existing issues and provide feedback or react to them.

With pull requests:
  - Open your pull request against `master`
  - It should pass all tests in the continuous integration pipeline (TravisCI).
  - You should add/modify tests to cover your proposed code changes.
  - If your pull request contains a new feature, please document it in this README.


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
