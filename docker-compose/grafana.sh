#!/bin/sh

# Exported dashboard can't be auto provisioned, fix dashboard before starting Grafana
# Set time range to 6h / 5s refresh
mkdir -p /var/lib/grafana/dashboards
cat /dashboard.json | sed -e 's/${DS_PROMETHEUS}/1/;s/now-7d/now-6h/;s/\"refresh\": \"1m\"/\"refresh\": \"5s\"/' > /var/lib/grafana/dashboards/dashboard.json
cat /dashboard-without-pulse.json | sed -e 's/${DS_PROMETHEUS}/1/' > /var/lib/grafana/dashboards/dashboard-without-pulse.json
cat /dashboard-prices.json | sed -e 's/${DS_TIBBER-EXPORTER}/2/' > /var/lib/grafana/dashboards/dashboard-prices.json
cat /dashboard-monthly.json | sed -e 's/${DS_PROMETHEUS}/1/' > /var/lib/grafana/dashboards/dashboard-monthly.json
/run.sh
