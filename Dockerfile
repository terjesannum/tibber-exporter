FROM golang:1.19.0-alpine as builder

WORKDIR /workspace
COPY go.* ./
RUN go mod download

COPY . /workspace

RUN CGO_ENABLED=0 go build -a -o tibber-exporter .

FROM alpine:3.16.2

LABEL org.opencontainers.image.title="tibber-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Tibber power usage and costs" \
      org.opencontainers.image.authors="Terje Sannum <terje@offpiste.org>" \
      org.opencontainers.image.url="https://github.com/terjesannum/tibber-exporter" \
      org.opencontainers.image.source="https://github.com/terjesannum/tibber-exporter"

WORKDIR /

COPY --from=builder /workspace/tibber-exporter .
USER 65532:65532

ENTRYPOINT ["/tibber-exporter"]
