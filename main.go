package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/terjesannum/tibber-exporter/internal/home"
	"github.com/terjesannum/tibber-exporter/internal/metrics"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
	"golang.org/x/exp/slices"
)

var (
	token                   string
	homesQuery              tibber.HomesQuery
	liveUrl                 string
	liveMeasurements        stringArgs
	disableLiveMeasurements stringArgs
	listenAddress           string
	userAgent               string
)

type (
	stringArgs []string
	transport  struct {
		Token     string
		UserAgent string
	}
)

func (sa *stringArgs) String() string {
	return fmt.Sprintln(*sa)
}

func (sa *stringArgs) Set(s string) error {
	*sa = append(*sa, s)
	return nil
}

func init() {
	flag.StringVar(&token, "token", os.Getenv("TIBBER_TOKEN"), "Tibber API token")
	flag.StringVar(&liveUrl, "live-url", "", "Override websocket url for live measurements")
	flag.Var(&liveMeasurements, "live", "Ids of homes to always start live measurements")
	flag.Var(&disableLiveMeasurements, "disable-live", "Ids of homes to disable live measurements")
	flag.StringVar(&listenAddress, "listen-address", ":8080", "Address to listen on for HTTP requests")
	flag.Parse()
	if userAgent == "" {
		userAgent = "tibber-exporter (https://github.com/terjesannum/tibber-exporter)"
	}
}

func exit(msg string) {
	log.Println(msg)
	time.Sleep(10 * time.Second)
	os.Exit(1)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.Token))
	req.Header.Set("User-Agent", t.UserAgent)
	return http.DefaultTransport.RoundTrip(req)
}

func getHomesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(homesQuery)
}

func main() {
	log.Printf("Starting %s\n", userAgent)
	ctx := context.Background()
	hc := &http.Client{Transport: &transport{Token: token, UserAgent: userAgent}}
	client := graphql.NewClient("https://api.tibber.com/v1-beta/gql", hc)

	err := client.Query(ctx, &homesQuery, nil)
	if err != nil {
		exit(fmt.Sprintf("Error getting homes: %v. Exiting...", err))
	}
	http.HandleFunc("/homes", getHomesHandler)
	wsUrl := homesQuery.Viewer.WebsocketSubscriptionUrl
	log.Printf("Websocket url: %s\n", wsUrl)
	if liveUrl != "" && liveUrl != wsUrl {
		log.Printf("Overiding websocket url with: %s\n", liveUrl)
		wsUrl = liveUrl
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
				string(s.Id),
				s.AppNickname,
				s.Address.Address1,
				s.Address.Address2,
				s.Address.Address3,
				s.Address.PostalCode,
				s.Address.City,
				s.Address.Country,
				s.Address.Latitude,
				s.Address.Longitude,
				s.TimeZone,
				s.CurrentSubscription.PriceInfo.Current.Currency,
			).Set(1)
			metrics.GridInfo.WithLabelValues(
				string(s.Id),
				s.MeteringPointData.GridCompany,
				s.MeteringPointData.PriceAreaCode,
			).Set(1)
			prometheus.MustRegister(metrics.NewHomeCollector(h))
			log.Printf("Realtime consumption enabled for %v: %v\n", s.Id, s.Features.RealTimeConsumptionEnabled)
			if (s.Features.RealTimeConsumptionEnabled || slices.Contains(liveMeasurements, string(s.Id))) && !slices.Contains(disableLiveMeasurements, string(s.Id)) {
				log.Printf("Starting live measurements monitoring of home %v\n", s.Id)
				go h.SubscribeMeasurements(ctx, hc, wsUrl, token)
				prometheus.MustRegister(metrics.NewMeasurementCollector(string(s.Id), &h.Measurements.LiveMeasurement, &h.TimestampedValues))
				started = append(started, string(s.Id))
			} else {
				log.Printf("Live measurements not available for home %v\n", s.Id)
			}
			h.UpdatePrices(ctx, client)
			http.HandleFunc(fmt.Sprintf("/homes/%s/prices", h.Id), h.GetPricesHandler)
			ticker := time.NewTicker(time.Minute)
			quit := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						h.UpdatePrices(ctx, client)
						h.UpdatePrevious(ctx, client, tibber.ResolutionHourly)
						h.UpdatePrevious(ctx, client, tibber.ResolutionDaily)
						if slices.Contains(started, string(h.Id)) {
							timeDiff := time.Now().Sub(h.Measurements.LiveMeasurement.Timestamp)
							if timeDiff.Minutes() > 1 {
								exit(fmt.Sprintf("No measurements received for home %s since %s. Exiting...\n", h.Id, h.Measurements.LiveMeasurement.Timestamp))
							}
						}
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
				exit(fmt.Sprintf("Monitoring of home %s not started. Exiting...\n", l))
			}
		}
	}

	log.Println("Starting http listener")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Tibber exporter")
	})
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(listenAddress, nil)
	exit(fmt.Sprintf("Error: %v", err))
}
