package metrics

import (
	"log"
	"time"

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
)

type MeasurementCollector struct {
	measurements     *tibber.LiveMeasurement
	consumption      *prometheus.Desc
	consumptionTotal *prometheus.Desc
	costTotal        *prometheus.Desc
}

func NewMeasurementCollector(homeId string, m *tibber.LiveMeasurement) *MeasurementCollector {
	return &MeasurementCollector{
		measurements: m,
		consumption: prometheus.NewDesc(
			"tibber_power_consumption",
			"Current power consumption",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
		consumptionTotal: prometheus.NewDesc(
			"tibber_power_consumption_day_total",
			"Total power consumption since midnight",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
		costTotal: prometheus.NewDesc(
			"tibber_power_cost_day_total",
			"Total power cost since midnight",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
	}
}

func (c *MeasurementCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.consumption
	ch <- c.consumptionTotal
	ch <- c.costTotal
}

func (c *MeasurementCollector) Collect(ch chan<- prometheus.Metric) {
	timeDiff := time.Now().Sub(c.measurements.Timestamp)
	if timeDiff.Minutes() > 5 {
		log.Printf("Measurements to old: %s\n", c.measurements.Timestamp)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		c.consumption,
		prometheus.GaugeValue,
		c.measurements.Power,
	)
	ch <- prometheus.MustNewConstMetric(
		c.consumptionTotal,
		prometheus.GaugeValue,
		c.measurements.AccumulatedConsumption,
	)
	ch <- prometheus.MustNewConstMetric(
		c.costTotal,
		prometheus.GaugeValue,
		c.measurements.AccumulatedCost,
	)
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
