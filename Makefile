build: deps
	CGO_ENABLED=0 go build -o bin/tibber-exporter \
		-ldflags " \
		-X github.com/prometheus/common/version.BuildUser=$(shell id -un) \
		-X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
		-X github.com/prometheus/common/version.Revision=$(shell git rev-parse HEAD) \
		-X github.com/prometheus/common/version.Version=$(shell sed -n 's/^version: //p' charts/tibber-exporter/Chart.yaml) \
		-X github.com/prometheus/common/version.BuildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
		"
deps:
	go mod download

