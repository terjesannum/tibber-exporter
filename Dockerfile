FROM golang:1.18.5-alpine as builder

WORKDIR /workspace
COPY go.* ./
RUN go mod download

COPY . /workspace

RUN CGO_ENABLED=0 go build -a -o tibber-exporter .

FROM alpine:3.15.5

LABEL org.opencontainers.image.authors="Terje Sannum <terje@offpiste.org>" \
      org.opencontainers.image.source="https://github.com/terjesannum/tibber-exporter"

WORKDIR /

COPY --from=builder /workspace/tibber-exporter .
USER 65532:65532

ENTRYPOINT ["/tibber-exporter"]
