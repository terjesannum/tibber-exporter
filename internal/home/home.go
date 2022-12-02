package home

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/hasura/go-graphql-client"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
)

type Measurements struct {
	LiveMeasurement tibber.LiveMeasurement `graphql:"liveMeasurement(homeId: $id)"`
}

type Home struct {
	Id                graphql.ID
	Prices            tibber.Prices `graphql:"home(id: $id)"`
	PreviousHour      tibber.PreviousPower
	PreviousDay       tibber.PreviousPower
	Measurements      Measurements
	TimestampedValues tibber.TimestampedValues
}

func New(id graphql.ID) *Home {
	return &Home{
		Id: id,
	}
}

func (h *Home) UpdatePrices(ctx context.Context, client *graphql.Client) {
	var prices tibber.Prices
	log.Printf("Updating prices for %v\n", h.Id)
	err := client.Query(ctx, &prices, map[string]interface{}{
		"id": h.Id,
	})
	if err != nil {
		log.Println(err)
		return
	}
	h.Prices = prices
}

func (h *Home) UpdatePrevious(ctx context.Context, client *graphql.Client, res tibber.EnergyResolution) {
	var prev tibber.PreviousQuery
	log.Printf("Updating %v consumption for %v\n", res, h.Id)
	err := client.Query(ctx, &prev, map[string]interface{}{
		"id":         h.Id,
		"resolution": res,
	})
	if err != nil {
		log.Println(err)
		return
	}
	values := &h.PreviousHour
	if res == tibber.ResolutionDaily {
		values = &h.PreviousDay
	}
	if len(prev.Viewer.Home.Consumption.Nodes) == 0 {
		log.Printf("Got 0 consumption nodes for %v\n", h.Id)
		values.Consumption = nil
		values.Cost = nil
		return
	}
	node := prev.Viewer.Home.Consumption.Nodes[0]
	now := time.Now()
	age := now.Sub(node.To).Hours()
	if res == tibber.ResolutionDaily {
		age = age / 24
	}
	values.Timestamp = now
	if age < 1 {
		values.Consumption = node.Consumption
		values.Cost = node.Cost
	} else {
		values.Consumption = nil
		values.Cost = nil
	}
}

func (h *Home) SubscribeMeasurements(ctx context.Context, hc *http.Client, wsUrl string, token string) {
	log.Printf("Creating measurements subscription for home %v\n", h.Id)
	subscriber := graphql.NewSubscriptionClient(wsUrl)
	subscriber.WithConnectionParams(map[string]interface{}{"token": token})
	subscriber.WithLog(log.Println)
	subscriber.WithRetryTimeout(time.Second * 5)
	subscriber.WithWebSocketOptions(graphql.WebsocketOptions{HTTPClient: hc})
	subscriber.OnConnected(func() {
		log.Printf("Measurements subscription for home %v connected\n", h.Id)
	})
	subscriber.OnDisconnected(func() {
		log.Printf("Measurements subscription for home %v disconnected\n", h.Id)
		log.Println("Exiting...")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	})
	subscriber.OnError(func(sc *graphql.SubscriptionClient, err error) error {
		log.Printf("OnError: %v\n", err)
		log.Println("Exiting...")
		time.Sleep(10 * time.Second)
		os.Exit(1)
		return err
	})
	defer subscriber.Close()
	subscriptionId, err := subscriber.Subscribe(
		&h.Measurements,
		map[string]interface{}{
			"id": h.Id,
		},
		func(dataValue []byte, errValue error) error {
			var m Measurements
			err := json.Unmarshal(dataValue, &m)
			if err != nil {
				log.Println(err)
			} else {
				h.Measurements.LiveMeasurement.Power = m.LiveMeasurement.Power
				h.Measurements.LiveMeasurement.MinPower = m.LiveMeasurement.MinPower
				h.Measurements.LiveMeasurement.MaxPower = m.LiveMeasurement.MaxPower
				h.Measurements.LiveMeasurement.AveragePower = m.LiveMeasurement.AveragePower
				// Each hour tibber seems to adjust readings (to official hourly reading?) and the accumulated values could be a bit lower that the previous.
				// This causes problems for prometheus counters, so skip those values.
				if m.LiveMeasurement.AccumulatedConsumption > h.Measurements.LiveMeasurement.AccumulatedConsumption ||
					m.LiveMeasurement.Timestamp.YearDay() != h.Measurements.LiveMeasurement.Timestamp.YearDay() {
					h.Measurements.LiveMeasurement.AccumulatedConsumption = m.LiveMeasurement.AccumulatedConsumption
				} else {
					log.Printf("Accumulated consumption lower than stored value: %f(%s) < %f(%s)\n",
						m.LiveMeasurement.AccumulatedConsumption, m.LiveMeasurement.Timestamp, h.Measurements.LiveMeasurement.AccumulatedConsumption, h.Measurements.LiveMeasurement.Timestamp)
				}
				if m.LiveMeasurement.AccumulatedCost > h.Measurements.LiveMeasurement.AccumulatedCost ||
					m.LiveMeasurement.Timestamp.YearDay() != h.Measurements.LiveMeasurement.Timestamp.YearDay() {
					h.Measurements.LiveMeasurement.AccumulatedCost = m.LiveMeasurement.AccumulatedCost
				} else {
					log.Printf("Accumulated cost lower than stored value: %f(%s) < %f(%s)\n",
						m.LiveMeasurement.AccumulatedCost, m.LiveMeasurement.Timestamp, h.Measurements.LiveMeasurement.AccumulatedCost, h.Measurements.LiveMeasurement.Timestamp)
				}
				if m.LiveMeasurement.CurrentL1 != nil {
					h.TimestampedValues.CurrentL1.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.CurrentL1.Value = *m.LiveMeasurement.CurrentL1
				}
				if m.LiveMeasurement.CurrentL2 != nil {
					h.TimestampedValues.CurrentL2.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.CurrentL2.Value = *m.LiveMeasurement.CurrentL2
				}
				if m.LiveMeasurement.CurrentL3 != nil {
					h.TimestampedValues.CurrentL3.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.CurrentL3.Value = *m.LiveMeasurement.CurrentL3
				}
				if m.LiveMeasurement.VoltagePhase1 != nil {
					h.TimestampedValues.VoltagePhase1.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.VoltagePhase1.Value = *m.LiveMeasurement.VoltagePhase1
				}
				if m.LiveMeasurement.VoltagePhase2 != nil {
					h.TimestampedValues.VoltagePhase2.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.VoltagePhase2.Value = *m.LiveMeasurement.VoltagePhase2
				}
				if m.LiveMeasurement.VoltagePhase3 != nil {
					h.TimestampedValues.VoltagePhase3.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.VoltagePhase3.Value = *m.LiveMeasurement.VoltagePhase3
				}
				if m.LiveMeasurement.SignalStrength != nil {
					h.TimestampedValues.SignalStrength.Timestamp = m.LiveMeasurement.Timestamp
					h.TimestampedValues.SignalStrength.Value = *m.LiveMeasurement.SignalStrength
				}
				h.Measurements.LiveMeasurement.Timestamp = m.LiveMeasurement.Timestamp
			}
			return err
		})
	if err != nil {
		log.Println(err)
	}
	log.Printf("Starting subscription %v for home %v\n", subscriptionId, h.Id)
	subscriber.Run()
	log.Printf("Ended subscription %v for home %v\n", subscriptionId, h.Id)
}
