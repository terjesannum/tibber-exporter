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

const maxAge = 5

type MeasurementCollector struct {
	measurements      *tibber.LiveMeasurement
	timestampedValues *tibber.TimestampedValues
	consumption       *prometheus.Desc
	consumptionTotal  *prometheus.Desc
	costTotal         *prometheus.Desc
	current           *prometheus.Desc
	voltage           *prometheus.Desc
	signalStrength    *prometheus.Desc
}

func NewMeasurementCollector(homeId string, m *tibber.LiveMeasurement, tv *tibber.TimestampedValues) *MeasurementCollector {
	return &MeasurementCollector{
		measurements:      m,
		timestampedValues: tv,
		consumption: prometheus.NewDesc(
			"tibber_power_consumption",
			"Power consumption",
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
		current: prometheus.NewDesc(
			"tibber_current",
			"Line current",
			[]string{"line"},
			prometheus.Labels{"home_id": homeId},
		),
		voltage: prometheus.NewDesc(
			"tibber_voltage",
			"Phase voltage",
			[]string{"phase"},
			prometheus.Labels{"home_id": homeId},
		),
		signalStrength: prometheus.NewDesc(
			"tibber_signal_strength",
			"Signal strength",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
	}
}

func (c *MeasurementCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.consumption
	ch <- c.consumptionTotal
	ch <- c.costTotal
	ch <- c.current
	ch <- c.voltage
	ch <- c.signalStrength
}

func (c *MeasurementCollector) Collect(ch chan<- prometheus.Metric) {
	timeDiff := time.Now().Sub(c.measurements.Timestamp)
	if timeDiff.Minutes() > maxAge {
		log.Printf("Measurements to old: %s\n", c.measurements.Timestamp)
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.consumption,
			prometheus.GaugeValue,
			c.measurements.Power,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.consumptionTotal,
			prometheus.GaugeValue,
			c.measurements.AccumulatedConsumption,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.costTotal,
			prometheus.GaugeValue,
			c.measurements.AccumulatedCost,
		),
	)
	timeDiff = time.Now().Sub(c.timestampedValues.CurrentL1.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.CurrentL1.Timestamp,
			prometheus.MustNewConstMetric(
				c.current,
				prometheus.GaugeValue,
				c.timestampedValues.CurrentL1.Value,
				"1",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.CurrentL2.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.CurrentL2.Timestamp,
			prometheus.MustNewConstMetric(
				c.current,
				prometheus.GaugeValue,
				c.timestampedValues.CurrentL2.Value,
				"2",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.CurrentL3.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.CurrentL3.Timestamp,
			prometheus.MustNewConstMetric(
				c.current,
				prometheus.GaugeValue,
				c.timestampedValues.CurrentL3.Value,
				"3",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.VoltagePhase1.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.VoltagePhase1.Timestamp,
			prometheus.MustNewConstMetric(
				c.voltage,
				prometheus.GaugeValue,
				c.timestampedValues.VoltagePhase1.Value,
				"1",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.VoltagePhase2.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.VoltagePhase2.Timestamp,
			prometheus.MustNewConstMetric(
				c.voltage,
				prometheus.GaugeValue,
				c.timestampedValues.VoltagePhase2.Value,
				"2",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.VoltagePhase3.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.VoltagePhase3.Timestamp,
			prometheus.MustNewConstMetric(
				c.voltage,
				prometheus.GaugeValue,
				c.timestampedValues.VoltagePhase3.Value,
				"3",
			),
		)
	}
	timeDiff = time.Now().Sub(c.timestampedValues.SignalStrength.Timestamp)
	if timeDiff.Minutes() < maxAge {
		ch <- prometheus.NewMetricWithTimestamp(
			c.timestampedValues.SignalStrength.Timestamp,
			prometheus.MustNewConstMetric(
				c.signalStrength,
				prometheus.GaugeValue,
				c.timestampedValues.SignalStrength.Value,
			),
		)
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
		price:      prometheus.NewDesc("tibber_power_price", "Power price", []string{"type"}, prometheus.Labels{"home_id": homeId}),
		priceLevel: prometheus.NewDesc("tibber_power_price_level", "Power price level", nil, prometheus.Labels{"home_id": homeId}),
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
