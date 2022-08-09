package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
)

var (
	HomeInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tibber_home_info",
			Help: "Home info",
		},
		[]string{
			"home_id",
			"name",
			"address1",
			"address2",
			"address3",
			"postal_code",
			"city",
			"country",
			"latitude",
			"longitude",
			"currency",
		})
	Consumption = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tibber_power_consumption",
			Help: "Current power consumption",
		},
		[]string{
			"home_id",
		})
)

type CounterCollector struct {
	counterDesc *prometheus.Desc
	value       *float64
}

func (c *CounterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.counterDesc
}

func (c *CounterCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.counterDesc,
		prometheus.CounterValue,
		*c.value,
	)
}

func NewCounterCollector(name string, help string, homeId string, value *float64) *CounterCollector {
	return &CounterCollector{
		counterDesc: prometheus.NewDesc(name, help, nil, prometheus.Labels{"home_id": homeId}),
		value:       value,
	}
}

type PriceCollector struct {
	homeId     string
	prices     *tibber.Prices
	price      *prometheus.Desc
	priceLevel *prometheus.Desc
}

func NewPriceCollector(homeId string, prices *tibber.Prices) *PriceCollector {
	log.Printf("Creating price collector for home %s\n", homeId)
	return &PriceCollector{
		homeId:     homeId,
		prices:     prices,
		price:      prometheus.NewDesc("tibber_power_price", "Current power price", []string{"type"}, prometheus.Labels{"home_id": homeId}),
		priceLevel: prometheus.NewDesc("tibber_power_price_level", "Current power price level", nil, prometheus.Labels{"home_id": homeId}),
	}
}

func (c *PriceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.price
	ch <- c.priceLevel
}

func (c *PriceCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.price,
		prometheus.GaugeValue,
		c.prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Total,
		"total",
	)
	ch <- prometheus.MustNewConstMetric(
		c.price,
		prometheus.GaugeValue,
		c.prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Energy,
		"energy",
	)
	ch <- prometheus.MustNewConstMetric(
		c.price,
		prometheus.GaugeValue,
		c.prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Tax,
		"tax",
	)
	ch <- prometheus.MustNewConstMetric(
		c.priceLevel,
		prometheus.GaugeValue,
		float64(tibber.PriceLevel[string(c.prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Level)]),
	)
}
