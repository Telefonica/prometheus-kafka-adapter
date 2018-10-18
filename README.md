# prometheus-kafka-adapter

[![Build Status](https://travis-ci.org/Telefonica/prometheus-kafka-adapter.svg?branch=master)](https://travis-ci.org/Telefonica/prometheus-kafka-adapter)

prometheus-kafka-adapter is a service which receives [Prometheus](https://github.com/prometheus) requests through [`remote_write`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), marshal into JSON and sends them into [Kafka](https://github.com/apache/kafka).


## motivation

We use `prometheus-kafka-adapter` internally at Telefonica for dumping Prometheus metrics into an object store in diferent clouds, through [Kafka](https://github.com/apache/kafka) and Kafka-Connect.


## output

It produces the following messages in a kafka topic:

```json
{
  "__timestamp__": 1234567890,
  "__value__": 9876543210,
  
  "__name__": "up",
  "job": "federation",

  "label1": "value1",
  "label2": "value2"
}
```

`__timestamp__` and `__value__` are reserved values, and can't be used as label names. `__name__` defines the name of the metric.


## configuration

It can be configured with the following environment variables:

- `LOG_LEVEL`: defines log level for [`logrus`](https://github.com/sirupsen/logrus), can be `debug`, `info`, `warn`, `error`, `fatal` or `panic`, defaults to `info`.
- `KAFKA_BROKER_LIST`: defines kafka endpoint and port, defaults to `kafka:9092`.
- `KAFKA_TOPIC`: defines kafka topic to be used, defaults to `metrics`.
- `PORT`: defines http port to listen, defaults to `8080`, used directly by [gin](https://github.com/gin-gonic/gin).
- `GIN_MODE`: manage [gin](https://github.com/gin-gonic/gin) debug logging, can be `debug` or `release`.


## contributing

With issues:
  - Use the search tool before opening a new issue.
  - Please provide source code and commit sha if you found a bug.
  - Review existing issues and provide feedback or react to them.

With pull requests:
  - Open your pull request against `master`
  - It should pass all tests in the available continuous integrations systems such as TravisCI.
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
