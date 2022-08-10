package main

import (
	"log"
	"net/http"
	"os"

	"context"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/terjesannum/tibber-exporter/internal/home"
	"github.com/terjesannum/tibber-exporter/internal/metrics"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
	"golang.org/x/oauth2"
)

var (
	homesQuery tibber.HomesQuery
)

func main() {
	ctx := context.Background()

	token := os.Getenv("TIBBER_TOKEN")
	oauth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}))
	tibber := graphql.NewClient("https://api.tibber.com/v1-beta/gql", oauth)

	err := tibber.Query(ctx, &homesQuery, nil)
	if err != nil {
		log.Printf("Error getting homes: %v", err)
		log.Println("Exiting...")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}

	for _, s := range homesQuery.Viewer.Homes {
		log.Printf("Found home: %v - %v\n", s.Id, s.AppNickname)
		if s.CurrentSubscription.PriceInfo.Current.Currency == "" {
			log.Printf("No subscription found for home %v\n", s.Id)
		} else {
			log.Printf("Starting monitoring of home: %v - %v\n", s.Id, s.AppNickname)
			h := home.New(s.Id)
			metrics.HomeInfo.WithLabelValues(
				s.Id.(string),
				string(s.AppNickname),
				string(s.Address.Address1),
				string(s.Address.Address2),
				string(s.Address.Address3),
				string(s.Address.PostalCode),
				string(s.Address.City),
				string(s.Address.Country),
				string(s.Address.Latitude),
				string(s.Address.Longitude),
				string(s.CurrentSubscription.PriceInfo.Current.Currency),
			).Set(1)
			prometheus.MustRegister(metrics.NewPriceCollector(s.Id.(string), &h.Prices))
			if s.Features.RealTimeConsumptionEnabled {
				log.Printf("Starting live measurements monitoring of home %v\n", s.Id)
				go h.SubscribeMeasurements(ctx, token)
				prometheus.MustRegister(metrics.NewMeasurementCollector(s.Id.(string), &h.Measurements.LiveMeasurement, &h.TimestampedValues))
			} else {
				log.Printf("Live measurements not available for home %v\n", s.Id)
			}
			h.UpdatePrices(ctx, tibber)
			ticker := time.NewTicker(time.Minute)
			quit := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						h.UpdatePrices(ctx, tibber)
					case <-quit:
						ticker.Stop()
						return
					}
				}
			}()
		}
	}

	log.Println("Starting http listener")
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":8080", nil)
	log.Printf("Error: %v", err)
	time.Sleep(10 * time.Second)
}
