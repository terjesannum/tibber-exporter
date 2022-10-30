package metrics

import (
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/terjesannum/tibber-exporter/internal/home"
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
	consumptionMin    *prometheus.Desc
	consumptionMax    *prometheus.Desc
	consumptionAvg    *prometheus.Desc
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
		consumptionMin: prometheus.NewDesc(
			"tibber_power_consumption_day_min",
			"Minimum power consumtion since midnight",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
		consumptionMax: prometheus.NewDesc(
			"tibber_power_consumption_day_max",
			"Maximum power consumtion since midnight",
			nil,
			prometheus.Labels{"home_id": homeId},
		),
		consumptionAvg: prometheus.NewDesc(
			"tibber_power_consumption_day_avg",
			"Average power consumtion since midnight",
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
	ch <- c.consumptionMin
	ch <- c.consumptionMax
	ch <- c.consumptionAvg
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
			c.consumptionMin,
			prometheus.GaugeValue,
			c.measurements.MinPower,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.consumptionMax,
			prometheus.GaugeValue,
			c.measurements.MaxPower,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.consumptionAvg,
			prometheus.GaugeValue,
			c.measurements.AveragePower,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.consumptionTotal,
			prometheus.CounterValue,
			c.measurements.AccumulatedConsumption,
		),
	)
	ch <- prometheus.NewMetricWithTimestamp(
		c.measurements.Timestamp,
		prometheus.MustNewConstMetric(
			c.costTotal,
			prometheus.CounterValue,
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

type HomeCollector struct {
	home                    *home.Home
	price                   *prometheus.Desc
	priceLevel              *prometheus.Desc
	previousHourConsumption *prometheus.Desc
	previousHourCost        *prometheus.Desc
	previousDayConsumption  *prometheus.Desc
	previousDayCost         *prometheus.Desc
}

func NewHomeCollector(home *home.Home) *HomeCollector {
	log.Printf("Creating home collector for home %s\n", home.Id)
	return &HomeCollector{
		home:                    home,
		price:                   prometheus.NewDesc("tibber_power_price", "Power price", []string{"type"}, prometheus.Labels{"home_id": string(home.Id)}),
		priceLevel:              prometheus.NewDesc("tibber_power_price_level", "Power price level", nil, prometheus.Labels{"home_id": string(home.Id)}),
		previousHourConsumption: prometheus.NewDesc("tibber_power_consumption_previous_hour", "Power consumption previous hour", nil, prometheus.Labels{"home_id": string(home.Id)}),
		previousHourCost:        prometheus.NewDesc("tibber_power_cost_previous_hour", "Power cost previous hour", nil, prometheus.Labels{"home_id": string(home.Id)}),
		previousDayConsumption:  prometheus.NewDesc("tibber_power_consumption_previous_day", "Power consumption previous day", nil, prometheus.Labels{"home_id": string(home.Id)}),
		previousDayCost:         prometheus.NewDesc("tibber_power_cost_previous_day", "Power cost previous day", nil, prometheus.Labels{"home_id": string(home.Id)}),
	}
}

func (c *HomeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.price
	ch <- c.priceLevel
	ch <- c.previousHourConsumption
	ch <- c.previousHourCost
	ch <- c.previousDayConsumption
	ch <- c.previousDayCost
}

func (c *HomeCollector) Collect(ch chan<- prometheus.Metric) {
	if c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Total != nil {
		ch <- prometheus.MustNewConstMetric(
			c.price,
			prometheus.GaugeValue,
			*c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Total,
			"total",
		)
	}
	if c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Energy != nil {
		ch <- prometheus.MustNewConstMetric(
			c.price,
			prometheus.GaugeValue,
			*c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Energy,
			"energy",
		)
	}
	if c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Tax != nil {
		ch <- prometheus.MustNewConstMetric(
			c.price,
			prometheus.GaugeValue,
			*c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Tax,
			"tax",
		)
	}
	if c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Level != nil {
		ch <- prometheus.MustNewConstMetric(
			c.priceLevel,
			prometheus.GaugeValue,
			float64(tibber.PriceLevel[string(*c.home.Prices.Viewer.Home.CurrentSubscription.PriceInfo.Current.Level)]),
		)
	}
	if c.home.PreviousHour.Consumption != nil {
		ch <- prometheus.NewMetricWithTimestamp(
			c.home.PreviousHour.Timestamp,
			prometheus.MustNewConstMetric(
				c.previousHourConsumption,
				prometheus.GaugeValue,
				*c.home.PreviousHour.Consumption,
			),
		)
	}
	if c.home.PreviousHour.Cost != nil {
		ch <- prometheus.NewMetricWithTimestamp(
			c.home.PreviousHour.Timestamp,
			prometheus.MustNewConstMetric(
				c.previousHourCost,
				prometheus.GaugeValue,
				*c.home.PreviousHour.Cost,
			),
		)
	}
	if c.home.PreviousDay.Consumption != nil {
		ch <- prometheus.NewMetricWithTimestamp(
			c.home.PreviousDay.Timestamp,
			prometheus.MustNewConstMetric(
				c.previousDayConsumption,
				prometheus.GaugeValue,
				*c.home.PreviousDay.Consumption,
			),
		)
	}
	if c.home.PreviousDay.Cost != nil {
		ch <- prometheus.NewMetricWithTimestamp(
			c.home.PreviousDay.Timestamp,
			prometheus.MustNewConstMetric(
				c.previousDayCost,
				prometheus.GaugeValue,
				*c.home.PreviousDay.Cost,
			),
		)
	}
}
