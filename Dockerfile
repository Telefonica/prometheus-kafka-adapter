FROM alpine:3.14 as build

WORKDIR /src/prometheus-kafka-adapter

COPY vendor .
COPY go.mod .
COPY go.sum .
ADD . /src/prometheus-kafka-adapter

RUN apk add --no-cache gcc musl-dev
RUN go test -tags musl,static,netgo -mod=vendor ./...
RUN go build -tags "musl,static,netgo" -mod=vendor -o /prometheus-kafka-adapter ./...

FROM alpine:3.14

COPY schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter
