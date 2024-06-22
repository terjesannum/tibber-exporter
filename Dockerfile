FROM golang:1.21.5-alpine3.19 AS builder

RUN apk --update add ca-certificates make git
RUN echo 'tibber:*:65532:' > /tmp/group && \
    echo 'tibber:*:65532:65532:tibber:/:/tibber-exporter' > /tmp/passwd

WORKDIR /workspace
COPY . /workspace

RUN make build

FROM scratch

LABEL org.opencontainers.image.title="tibber-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Tibber power usage and costs" \
      org.opencontainers.image.authors="Terje Sannum <terje@offpiste.org>" \
      org.opencontainers.image.url="https://github.com/terjesannum/tibber-exporter" \
      org.opencontainers.image.source="https://github.com/terjesannum/tibber-exporter"

WORKDIR /
EXPOSE 8080

COPY --from=builder /tmp/passwd /tmp/group /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /workspace/tibber-exporter .

USER 65532:65532

ENTRYPOINT ["/tibber-exporter"]
