FROM alpine:3.13
ARG binary=prometheus-kafka-adapter

COPY schemas/metric.avsc /schemas/metric.avsc
COPY ${binary} /prometheus-kafka-adapter

CMD /prometheus-kafka-adapter
