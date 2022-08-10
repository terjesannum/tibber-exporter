package home

import (
	"encoding/json"
	"log"
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
	Id             graphql.ID
	Prices         tibber.Prices `graphql:"home(id: $id)"`
	Measurements   Measurements
	queryVariables map[string]interface{}
}

func New(id graphql.ID) *Home {
	h := &Home{
		Id: id,
		queryVariables: map[string]interface{}{
			"id": id,
		},
	}
	return h
}

func (h *Home) UpdatePrices(ctx context.Context, client *graphql.Client) {
	log.Printf("Updating prices for %v\n", h.Id)
	err := client.Query(ctx, &h.Prices, h.queryVariables)
	if err != nil {
		log.Println(err)
	}
}

func (h *Home) SubscribeMeasurements(ctx context.Context, token string) {
	log.Printf("Creating measurements subscription for home %v\n", h.Id)
	subscriber := graphql.NewSubscriptionClient("wss://api.tibber.com/v1-beta/gql/subscriptions").WithConnectionParams(map[string]interface{}{
		"token": token,
	}).WithLog(log.Println)
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
	subscriptionId, err := subscriber.Subscribe(&h.Measurements, h.queryVariables, func(dataValue []byte, errValue error) error {
		var m Measurements
		err := json.Unmarshal(dataValue, &m)
		if err != nil {
			log.Println(err)
		} else {
			h.Measurements.LiveMeasurement.Power = m.LiveMeasurement.Power
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
