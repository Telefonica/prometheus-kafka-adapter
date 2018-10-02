FROM golang:1.11-alpine as build

RUN apk add --no-cache alpine-sdk librdkafka librdkafka-dev

WORKDIR /src/prometheus-kafka-adapter
ADD . /src/prometheus-kafka-adapter

RUN go build -o /prometheus-kafka-adapter

FROM alpine
COPY --from=build /prometheus-kafka-adapter /
CMD prometheus-kafka-adapter