FROM alpine:3.14
ARG binary=prometheus-kafka-adapter

RUN apk add --no-cache ca-certificates
COPY prometheus-kafka-adapter-musl /${binary}
CMD /${binary}
