package tibber

import (
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

type Prices struct {
	Viewer struct {
		Home struct {
			CurrentSubscription struct {
				PriceInfo struct {
					Current struct {
						Total  *float64
						Energy *float64
						Tax    *float64
						Level  *string
					}
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
