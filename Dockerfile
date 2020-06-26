FROM golang:1.14.4-alpine3.11 as build

# Get prebuilt libkafka.
# XXX stop using the edgecommunity channel once librdkafka 1.3.0+ is officially published
RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edgecommunity http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk add --no-cache alpine-sdk 'librdkafka@edgecommunity>=1.3.0' 'librdkafka-dev@edgecommunity>=1.3.0'

WORKDIR /src/prometheus-kafka-adapter
ADD . /src/prometheus-kafka-adapter

RUN go test
RUN go build -o /prometheus-kafka-adapter

FROM alpine:3.11

RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edgecommunity http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk add --no-cache 'librdkafka@edgecommunity>=1.3.0'

COPY --from=build /src/prometheus-kafka-adapter/schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
