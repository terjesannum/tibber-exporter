package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"context"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/terjesannum/tibber-exporter/internal/home"
	"github.com/terjesannum/tibber-exporter/internal/metrics"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

var (
	homesQuery       tibber.HomesQuery
	liveMeasurements []string
)

func init() {
	var s string
	// Initialize with homes having live measurements. The data received in features.realTimeConsumptionEnabled is not always correct
	flag.StringVar(&s, "live", os.Getenv("TIBBER_LIVE_MEASUREMENTS"), "Comma separated list of homes with live measurements")
	flag.Parse()
	if s != "" {
		liveMeasurements = strings.Split(s, ",")
	}
}

func main() {
	ctx := context.Background()

	token := os.Getenv("TIBBER_TOKEN")
	oauth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}))
	client := graphql.NewClient("https://api.tibber.com/v1-beta/gql", oauth)

	err := client.Query(ctx, &homesQuery, nil)
	if err != nil {
		log.Printf("Error getting homes: %v", err)
		log.Println("Exiting...")
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}

	var started []string
	for _, s := range homesQuery.Viewer.Homes {
		log.Printf("Found home: %v - %v\n", s.Id, s.AppNickname)
		if s.CurrentSubscription.Id == nil {
			log.Printf("No subscription found for home %v\n", s.Id)
		} else {
			log.Printf("Starting monitoring of home: %v - %v\n", s.Id, s.AppNickname)
			log.Printf("Current subscription: %v\n", *s.CurrentSubscription.Id)
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
			prometheus.MustRegister(metrics.NewHomeCollector(h))
			if s.Features.RealTimeConsumptionEnabled || slices.Contains(liveMeasurements, s.Id.(string)) {
				log.Printf("Starting live measurements monitoring of home %v\n", s.Id)
				go h.SubscribeMeasurements(ctx, token)
				prometheus.MustRegister(metrics.NewMeasurementCollector(s.Id.(string), &h.Measurements.LiveMeasurement, &h.TimestampedValues))
				started = append(started, s.Id.(string))
			} else {
				log.Printf("Live measurements not available for home %v\n", s.Id)
			}
			h.UpdatePrices(ctx, client)
			ticker := time.NewTicker(time.Minute)
			quit := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						h.UpdatePrices(ctx, client)
						h.UpdatePrevious(ctx, client, tibber.ResolutionHourly)
						h.UpdatePrevious(ctx, client, tibber.ResolutionDaily)
					case <-quit:
						ticker.Stop()
						return
					}
				}
			}()
		}
	}

	// Exit if live monitoring of configured home for some reason hasn't started
	if len(liveMeasurements) > 0 {
		for _, l := range liveMeasurements {
			if !slices.Contains(started, l) {
				log.Printf("Monitoring of home %s not started. Exiting...\n", l)
				time.Sleep(10 * time.Second)
				os.Exit(1)
			}
		}
	}

	log.Println("Starting http listener")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Tibber prometheus exporter")
	})
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":8080", nil)
	log.Printf("Error: %v", err)
	time.Sleep(10 * time.Second)
}