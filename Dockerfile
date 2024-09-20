FROM golang:1.22.3-alpine as build
WORKDIR /src/prometheus-kafka-adapter

COPY go.mod go.sum vendor *.go ./

ADD . /src/prometheus-kafka-adapter

RUN apk add --no-cache gcc musl-dev
RUN go build -ldflags='-w -s -extldflags "-static"' -tags musl,static,netgo -mod=vendor -o /prometheus-kafka-adapter

FROM alpine:3.20

COPY schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
