FROM golang:1.13.6-alpine3.10 as build

RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
RUN echo "@edgecommunity http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories  
RUN apk add --no-cache alpine-sdk 'librdkafka@edgecommunity>=1.3.0' 'librdkafka-dev@edgecommunity>=1.3.0'

WORKDIR /src/prometheus-kafka-adapter
ADD . /src/prometheus-kafka-adapter

RUN go build -o /prometheus-kafka-adapter

FROM alpine:3.10

RUN apk add --no-cache librdkafka

COPY --from=build /src/prometheus-kafka-adapter/schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
