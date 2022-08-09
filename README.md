# Tibber prometheus exporter

Monitor your power usage and costs with Prometheus and Grafana.

![Grafana dashboard](grafana/dashboard.png)

## Description

This prometheus exporter will connect to the Tibber API, subscribe to updates from [Tibber Pulse](https://tibber.com/no/pulse) and make the metrics available for Prometheus.
See the provided [Grafana dashboard](grafana/dashboard.json) for examples on how they can be used.  
Note that the consumption, cost and price heatmap panels requires the [Hourly heatmap plugin](https://grafana.com/grafana/plugins/marcusolsson-hourly-heatmap-panel/).

If you don't have Tibber Pulse, only the power price metrics will be available.

## Running

Docker image is available on [ghcr.io](https://github.com/terjesannum/tibber-exporter/pkgs/container/tibber-exporter).

```sh
docker run -d -p 8080:8080 -e TIBBER_TOKEN=... --restart always ghcr.io/terjesannum/tibber-exporter:1
```

Go to [developer.tibber.com](https://developer.tibber.com/) to find your `TIBBER_TOKEN`.

## Metrics

```
# HELP tibber_home_info Home info
# TYPE tibber_home_info gauge
tibber_home_info{address1="Bedringens vei 1",address2="",address3="",city="OSLO",country="NO",currency="NOK",home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",latitude="59.9368932",longitude="10.736951",name="My home",postal_code="0450"} 1
# HELP tibber_power_consumption Current power consumption
# TYPE tibber_power_consumption gauge
tibber_power_consumption{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 47
# HELP tibber_power_consumption_day_total Total power consumption since midnight
# TYPE tibber_power_consumption_day_total counter
tibber_power_consumption_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 2.995177
# HELP tibber_power_cost_day_total Total power cost since midnight
# TYPE tibber_power_cost_day_total counter
tibber_power_cost_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 10.541825
# HELP tibber_power_price Current power price
# TYPE tibber_power_price gauge
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="energy"} 2.804
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="tax"} 0.711
tibber_power_price{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de",type="total"} 3.515
# HELP tibber_power_price_level Current power price level
# TYPE tibber_power_price_level gauge
tibber_power_price_level{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 4
```

### Consider becoming a Tibber customer?

I would be happy if you use [this referral code](https://invite.tibber.com/qandobma). That will give each of us a bonus to use on stuff like Tibber Pulse :smile:
