FROM alpine:3.13
ARG binary=prometheus-kafka-adapter

RUN apk add --no-cache ca-certificates
COPY schemas/metric.avsc /schemas/metric.avsc
COPY prometheus-kafka-adapter-musl /${binary}

CMD /${binary}
