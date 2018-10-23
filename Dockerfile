FROM golang:1.11-alpine as build

RUN apk add --no-cache alpine-sdk librdkafka librdkafka-dev

WORKDIR /src/prometheus-kafka-adapter
ADD . /src/prometheus-kafka-adapter

RUN go build -o /prometheus-kafka-adapter

FROM alpine:3.8

RUN apk add --no-cache librdkafka

COPY --from=build /src/prometheus-kafka-adapter/schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
