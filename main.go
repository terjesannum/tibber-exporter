package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/prometheus/client_golang/prometheus"
	promver "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/terjesannum/tibber-exporter/internal/home"
	"github.com/terjesannum/tibber-exporter/internal/metrics"
	"github.com/terjesannum/tibber-exporter/internal/tibber"
	"golang.org/x/exp/slices"
)

var (
	token                    string
	homesQuery               tibber.HomesQuery
	liveFeedTimeout          int
	liveUrl                  string
	liveMeasurements         stringArgs
	disableLiveMeasurements  stringArgs
	disableSubscriptionCheck bool
	showVersion              bool
	listenAddress            string
	userAgent                string
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
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.StringVar(&token, "token", os.Getenv("TIBBER_TOKEN"), "Tibber API token")
	flag.IntVar(&liveFeedTimeout, "live-feed-timeout", 1, "Timeout in minutes for live feed")
	flag.StringVar(&liveUrl, "live-url", "", "Override websocket url for live measurements")
	flag.Var(&liveMeasurements, "live", "Ids of homes to always start live measurements")
	flag.Var(&disableLiveMeasurements, "disable-live", "Ids of homes to disable live measurements")
	flag.BoolVar(&disableSubscriptionCheck, "disable-subscription-check", false, "Disable check on active Tibber subscription")
	flag.StringVar(&listenAddress, "listen-address", ":8080", "Address to listen on for HTTP requests")
	flag.Parse()
	if version.Version == "" {
		userAgent = "tibber-exporter (https://github.com/terjesannum/tibber-exporter)"
	} else {
		userAgent = fmt.Sprintf("tibber-exporter/%s (https://github.com/terjesannum/tibber-exporter)", version.Version)
	}

	//Get homes with Forcefully enabled/disabled live Measurement Feature from Environment Variable and append them to slice
	envEnabledHomes, envPresent := os.LookupEnv("TIBBER_LIVE_ENABLE_HOMES")
	if envPresent {
		envEnabledHomes := strings.Split(envEnabledHomes, ",")
		for _, enabledHome := range envEnabledHomes {
			if len(enabledHome) >= 1 {
				liveMeasurements = append(liveMeasurements, enabledHome)
			}
		}
	}
	envDisabledHomes, envPresent := os.LookupEnv("TIBBER_LIVE_DISABLE_HOMES")
	if envPresent {
		envDisabledHomes := strings.Split(envDisabledHomes, ",")
		for _, disabledHome := range envDisabledHomes {
			if len(disabledHome) >= 1 {
				disableLiveMeasurements = append(disableLiveMeasurements, disabledHome)
			}
		}
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
	if showVersion {
		fmt.Printf("%s\n", version.Print("tibber-exporter"))
		os.Exit(0)
	}
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
	prometheus.MustRegister(promver.NewCollector("tibber_exporter"))
	var started []string
	for _, s := range homesQuery.Viewer.Homes {
		s := s
		log.Printf("Found home: %v - %v\n", s.Id, s.AppNickname)
		if s.CurrentSubscription.Id == nil {
			log.Printf("No subscription found for home %v\n", s.Id)
		}
		if s.CurrentSubscription.Id != nil || disableSubscriptionCheck {
			log.Printf("Starting monitoring of home: %v - %v\n", s.Id, s.AppNickname)
			if s.CurrentSubscription.Id == nil {
				log.Println("Current subscription: n/a")
			} else {
				log.Printf("Current subscription: %v\n", *s.CurrentSubscription.Id)
			}
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
				log.Printf("Live feed timeout: %v minute\n", liveFeedTimeout)
				go h.SubscribeMeasurements(ctx, hc, wsUrl, token)
				prometheus.MustRegister(metrics.NewMeasurementCollector(string(s.Id), &h.Measurements.LiveMeasurement, &h.TimestampedValues, &h.GaugeValues))
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
						if s.CurrentSubscription.Id != nil {
							h.UpdatePrices(ctx, client)
						}
						h.UpdatePrevious(ctx, client, tibber.ResolutionHourly)
						h.UpdatePrevious(ctx, client, tibber.ResolutionDaily)
						if slices.Contains(started, string(h.Id)) {
							timeDiff := time.Now().Sub(h.Measurements.LiveMeasurement.Timestamp)
							if timeDiff.Minutes() > float64(liveFeedTimeout) {
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

	log.Printf("Starting http listener %s\n", listenAddress)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Tibber exporter")
	})
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(listenAddress, nil)
	exit(fmt.Sprintf("Error: %v", err))
}
