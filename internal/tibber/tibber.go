package tibber

import (
	"encoding/json"
	"time"

	"github.com/hasura/go-graphql-client"
)

type HomesQuery struct {
	Viewer struct {
		WebsocketSubscriptionUrl string
		Homes                    []struct {
			Id          graphql.ID
			AppNickname string
			Address     struct {
				Address1   string
				Address2   string
				Address3   string
				City       string
				PostalCode string
				Country    string
				Latitude   string
				Longitude  string
			}
			TimeZone          string
			MeteringPointData struct {
				GridCompany   string
				PriceAreaCode string
			}
			CurrentSubscription struct {
				Id        *graphql.ID
				PriceInfo struct {
					Current struct {
						Currency string
					}
				}
			}
			Features struct {
				RealTimeConsumptionEnabled bool
			}
		}
	}
}

func (hq HomesQuery) MarshalJSON() ([]byte, error) {
	type homeJSON struct {
		Id            graphql.ID `json:"homeId"`
		AppNickname   string     `json:"name,omitempty"`
		Address1      string     `json:"address1,omitempty"`
		Address2      string     `json:"address2,omitempty"`
		Address3      string     `json:"address3,omitempty"`
		City          string     `json:"city,omitempty"`
		PostalCode    string     `json:"postalCode,omitempty"`
		Country       string     `json:"country,omitempty"`
		Latitude      string     `json:"latitude,omitempty"`
		Longitude     string     `json:"longitude,omitempty"`
		TimeZone      string     `json:"timezone,omitempty"`
		GridCompany   string     `json:"gridCompany,omitempty"`
		PriceAreaCode string     `json:"priceAreaCode,omitempty"`
		Currency      string     `json:"currency,omitempty"`
	}
	var homes []homeJSON
	for _, home := range hq.Viewer.Homes {
		if home.CurrentSubscription.Id != nil {
			hj := homeJSON{
				Id:            home.Id,
				AppNickname:   home.AppNickname,
				Address1:      home.Address.Address1,
				Address2:      home.Address.Address2,
				Address3:      home.Address.Address3,
				City:          home.Address.City,
				PostalCode:    home.Address.PostalCode,
				Country:       home.Address.Country,
				Latitude:      home.Address.Latitude,
				Longitude:     home.Address.Longitude,
				TimeZone:      home.TimeZone,
				GridCompany:   home.MeteringPointData.GridCompany,
				PriceAreaCode: home.MeteringPointData.PriceAreaCode,
				Currency:      home.CurrentSubscription.PriceInfo.Current.Currency,
			}
			homes = append(homes, hj)
		}
	}
	return json.Marshal(homes)
}

type Price struct {
	StartsAt *time.Time
	Total    *float64
	Energy   *float64
	Tax      *float64
	Level    *string
}

// Return price level as int also in json so value mappings in Grafana will be the same as for Prometheus metrics
func (p Price) MarshalJSON() ([]byte, error) {
	type priceJSON struct {
		StartsAt *time.Time `json:"startsAt"`
		Total    *float64   `json:"total"`
		Energy   *float64   `json:"energy"`
		Tax      *float64   `json:"tax"`
		Level    int        `json:"level"`
	}
	pj := priceJSON{
		StartsAt: p.StartsAt,
		Total:    p.Total,
		Energy:   p.Energy,
		Tax:      p.Tax,
		Level:    PriceLevel[*p.Level],
	}
	return json.Marshal(pj)
}

type Prices struct {
	Viewer struct {
		Home struct {
			CurrentSubscription struct {
				PriceInfo struct {
					Current  Price
					Today    []Price
					Tomorrow []Price
				}
			}
		} `graphql:"home(id: $id)"`
	}
}

type LiveMeasurement struct {
	Timestamp               time.Time
	Power                   float64
	MinPower                float64
	MaxPower                float64
	AveragePower            float64
	LastMeterConsumption    float64
	AccumulatedConsumption  float64
	AccumulatedCost         float64
	CurrentL1               *float64
	CurrentL2               *float64
	CurrentL3               *float64
	VoltagePhase1           *float64
	VoltagePhase2           *float64
	VoltagePhase3           *float64
	SignalStrength          *float64
	AccumulatedProduction   float64
	AccumulatedReward       *float64
	PowerProduction         float64
	PowerReactive           *float64
	PowerProductionReactive *float64
	MinPowerProduction      float64
	MaxPowerProduction      float64
	LastMeterProduction     float64
	PowerFactor             *float64
}

var PriceLevel = map[string]int{
	"VERY_CHEAP":     1,
	"CHEAP":          2,
	"NORMAL":         3,
	"EXPENSIVE":      4,
	"VERY_EXPENSIVE": 5,
}

// Keep timestamp for values not present in every live measurement reading
type timestampedValue struct {
	Timestamp time.Time
	Value     float64
}

type TimestampedValues struct {
	CurrentL1               timestampedValue
	CurrentL2               timestampedValue
	CurrentL3               timestampedValue
	VoltagePhase1           timestampedValue
	VoltagePhase2           timestampedValue
	VoltagePhase3           timestampedValue
	SignalStrength          timestampedValue
	PowerReactive           timestampedValue
	PowerProductionReactive timestampedValue
	PowerFactor             timestampedValue
}

type EnergyResolution string

const ResolutionHourly EnergyResolution = "HOURLY"
const ResolutionDaily EnergyResolution = "DAILY"

type PreviousQuery struct {
	Viewer struct {
		Home struct {
			Consumption struct {
				Nodes []struct {
					From        time.Time
					To          time.Time
					Consumption *float64
					Cost        *float64
				}
			} `graphql:"consumption(resolution: $resolution, last: 1)"`
			Production struct {
				Nodes []struct {
					From       time.Time
					To         time.Time
					Production *float64
					Profit     *float64
				}
			} `graphql:"production(resolution: $resolution, last: 1)"`
		} `graphql:"home(id: $id)"`
	}
}

type PreviousPower struct {
	Timestamp   time.Time
	Consumption *float64
	Cost        *float64
	Production  *float64
	Profit      *float64
}
