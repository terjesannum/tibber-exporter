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
docker run -d -p 8080:8080 -e TIBBER_TOKEN=... --restart always ghcr.io/terjesannum/tibber-exporter:3
```

Go to [developer.tibber.com](https://developer.tibber.com/) to find your `TIBBER_TOKEN`.

## Metrics

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
# HELP tibber_power_consumption_day_total Total power consumption since midnight
# TYPE tibber_power_consumption_day_total counter
tibber_power_consumption_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 3.313674 1660135437500
# HELP tibber_power_cost_day_total Total power cost since midnight
# TYPE tibber_power_cost_day_total counter
tibber_power_cost_day_total{home_id="69e3138e-8a89-43d3-8179-f5e1cb2199de"} 10.981923 1660135437500
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
