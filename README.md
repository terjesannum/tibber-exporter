# Tibber prometheus exporter

Monitor your power usage and costs with Prometheus and Grafana.

![Grafana dashboard](grafana/dashboard.png)

## Description

This prometheus exporter will connect to the Tibber API, subscribe to updates from [Tibber Pulse](https://tibber.com/no/pulse) or [Watty](https://tibber.com/se/store/produkt/watty-tibber) and make the metrics available for Prometheus.

See the provided [Grafana dashboard](grafana/dashboard.json) for examples on how they can be used.  
Note that the consumption, cost and price heatmap panels requires the [Hourly heatmap plugin](https://grafana.com/grafana/plugins/marcusolsson-hourly-heatmap-panel/).

#### Don't have Tibber Pulse or Watty?

If you don't have a device for live measurements, only the power price metrics will be available on that dashboard. It is possible to create a dashboard with historic consumption and cost using the `..._previous_day` and `..._previous_hour` metrics, but the availability of those metrics may vary between grid companies. See the [dashboard without pulse](grafana/dashboard-without-pulse.json) for an example using the previous day metrics:

![Grafana dashboard without pulse](grafana/dashboard-without-pulse.png)


## Running

Docker image is available on [ghcr.io](https://github.com/terjesannum/tibber-exporter/pkgs/container/tibber-exporter).

A `TIBBER_TOKEN` is required to use the Tibber API, go to [developer.tibber.com](https://developer.tibber.com/) to find yours.

Regardless of how you run this program, it is important to run it with an automatic restart mechanism. If the live feed from Tibber for some reason is interrupted or not available, the program will take a short pause and exit. The pause is to avoid a restart loop and trigger rate limiting in the Tibber API, and just exiting and restarting is better than trying to handle every possible error situation.

### Docker compose

If you don't already run Grafana and Prometheus, you can try a complete setup with `docker compose`.

```sh
cd docker-compose
TIBBER_TOKEN=... docker compose up
```

Then go to http://localhost:3000/ and find the dashboards in the General folder.

### Kubernetes

Install in your kubernetes cluster with [Helm](https://helm.sh/). First add the the helm repository:

```
helm repo add tibber-exporter https://terjesannum.github.io/tibber-exporter/
helm repo update
```

Then install the helm chart:

```
helm install tibber-exporter tibber-exporter/tibber-exporter --set tibberToken=...
```

This with install the exporter with the `prometheus.io/scrape` annotation set to `true`. If you run the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator), install with `serviceMonitor.enabled=true` to create a `ServiceMonitor` instead:

```
helm install tibber-exporter tibber-exporter/tibber-exporter --set tibberToken=... --set serviceMonitor.enabled=true
```

### Docker container

```sh
docker run -d -p 8080:8080 -e TIBBER_TOKEN=... --restart always ghcr.io/terjesannum/tibber-exporter:latest
```

## Prometheus

Prometheus need to be configured to scrape the exporter, so add a scrape job to `/etc/prometheus/prometheus.yml`:
```
scrape_configs:

  - job_name: "tibber-exporter"
    scrape_interval: 5s
    static_configs:
      - targets: ["localhost:8080"]
```

How often the data is updated depends on your energy meter. Look at the logs from the exporter to see how often it receives updates, and adjust the scrape interval. Shorter scrape interval generates more data, so consider scraping less frequent if you only use a dashboard with a wide time range and don't need "live" updates.

Also remember that Prometheus is designed for monitoring and not precise calculation, so don't expect the result of the queries to excactly match your electricity bill.

## Grafana

Import the [dashboard](grafana/dashboard.json) or use id `16804` and import from [grafana.com](https://grafana.com/grafana/dashboards/16804-tibber/). Then select the Prometheus datasource that scrapes the exporter.

## Metrics

Example of metrics provided by this exporter:

```
# HELP tibber_current Line current
# TYPE tibber_current gauge
tibber_current{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",line="1"} 0.8 1660135430000
tibber_current{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",line="3"} 0.4 1660135430000
# HELP tibber_home_info Home info
# TYPE tibber_home_info gauge
tibber_home_info{address1="Bedringens vei 1",address2="",address3="",city="OSLO",country="NO",currency="NOK",home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",latitude="59.9368932",longitude="10.736951",name="My home",postal_code="0450"} 1
# HELP tibber_power_consumption Power consumption
# TYPE tibber_power_consumption gauge
tibber_power_consumption{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 47 1660135437500
# HELP tibber_power_consumption_day_avg Average power consumtion since midnight
# TYPE tibber_power_consumption_day_avg gauge
tibber_power_consumption_day_avg{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 218.7 1660135437500
# HELP tibber_power_consumption_day_max Maximum power consumtion since midnight
# TYPE tibber_power_consumption_day_max gauge
tibber_power_consumption_day_max{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 2116 1660135437500
# HELP tibber_power_consumption_day_min Minimum power consumtion since midnight
# TYPE tibber_power_consumption_day_min gauge
tibber_power_consumption_day_min{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 44 1660135437500
# HELP tibber_power_consumption_day_total Total power consumption since midnight
# TYPE tibber_power_consumption_day_total counter
tibber_power_consumption_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 3.313674 1660135437500
# HELP tibber_power_consumption_previous_day Power consumption previous day
# TYPE tibber_power_consumption_previous_day gauge
tibber_power_consumption_previous_day{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 29.741 1660135437500
# HELP tibber_power_consumption_previous_hour Power consumption previous hour
# TYPE tibber_power_consumption_previous_hour gauge
tibber_power_consumption_previous_hour{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 1.82 1660135437500
# HELP tibber_power_cost_day_total Total power cost since midnight
# TYPE tibber_power_cost_day_total counter
tibber_power_cost_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 10.981923 1660135437500
# HELP tibber_power_cost_previous_day Power cost previous day
# TYPE tibber_power_cost_previous_day gauge
tibber_power_cost_previous_day{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 133.901981625 1660135437500
# HELP tibber_power_cost_previous_hour Power cost previous hour
# TYPE tibber_power_cost_previous_hour gauge
tibber_power_cost_previous_hour{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 7.37839375 1660135437500
# HELP tibber_power_price Power price
# TYPE tibber_power_price gauge
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="energy"} 2.5294
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="tax"} 0.6423
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="total"} 3.1717
# HELP tibber_power_price_level Power price level
# TYPE tibber_power_price_level gauge
tibber_power_price_level{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 3
# HELP tibber_signal_strength Signal strength
# TYPE tibber_signal_strength gauge
tibber_signal_strength{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} -60 1660135347500
# HELP tibber_voltage Phase voltage
# TYPE tibber_voltage gauge
tibber_voltage{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",phase="1"} 234.6 1660135430000
tibber_voltage{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",phase="2"} 234.1 1660135430000
tibber_voltage{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",phase="3"} 234.9 1660135430000
```

### Consider becoming a Tibber customer?

I would be happy if you use [this referral code](https://invite.tibber.com/qandobma). That will give each of us a bonus to use on stuff like Tibber Pulse :smile:

Or register the code `qandobma` directly in the app if you signed up without an invite.
