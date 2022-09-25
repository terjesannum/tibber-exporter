package tibber

import (
	"time"

	"github.com/hasura/go-graphql-client"
)

type HomesQuery struct {
	Viewer struct {
		Homes []struct {
			Id          graphql.ID
			AppNickname graphql.String
			Address     struct {
				Address1   graphql.String
				Address2   graphql.String
				Address3   graphql.String
				City       graphql.String
				PostalCode graphql.String
				Country    graphql.String
				Latitude   graphql.String
				Longitude  graphql.String
			}
			CurrentSubscription struct {
				PriceInfo struct {
					Current struct {
						Currency graphql.String
					}
				}
			}
			Features struct {
				RealTimeConsumptionEnabled graphql.Boolean
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
						Total  float64
						Energy float64
						Tax    float64
						Level  graphql.String
					}
				}
			}
		} `graphql:"home(id: $id)"`
	}
}

type LiveMeasurement struct {
	Timestamp              time.Time
	Power                  float64
	MinPower               float64
	MaxPower               float64
	AveragePower           float64
	AccumulatedConsumption float64
	AccumulatedCost        float64
	CurrentL1              *float64
	CurrentL2              *float64
	CurrentL3              *float64
	VoltagePhase1          *float64
	VoltagePhase2          *float64
	VoltagePhase3          *float64
	SignalStrength         *float64
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
	CurrentL1      timestampedValue
	CurrentL2      timestampedValue
	CurrentL3      timestampedValue
	VoltagePhase1  timestampedValue
	VoltagePhase2  timestampedValue
	VoltagePhase3  timestampedValue
	SignalStrength timestampedValue
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
		} `graphql:"home(id: $id)"`
	}
}

type PreviousPower struct {
	Timestamp   time.Time
	Consumption *float64
	Cost        *float64
}
