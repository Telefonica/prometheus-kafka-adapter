FROM golang:1.14.4-alpine3.12 as build

# Get prebuilt libkafka.
RUN apk add --no-cache alpine-sdk 'librdkafka>=1.3.0' 'librdkafka-dev>=1.3.0'

WORKDIR /src/prometheus-kafka-adapter
ADD . /src/prometheus-kafka-adapter

RUN go test
RUN go build -o /prometheus-kafka-adapter

FROM alpine:3.12

RUN apk add --no-cache 'librdkafka>=1.3.0'

COPY --from=build /src/prometheus-kafka-adapter/schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
