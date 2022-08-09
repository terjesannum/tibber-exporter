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
	AccumulatedConsumption float64
	AccumulatedCost        float64
}

var PriceLevel = map[string]int{
	"VERY_CHEAP":     1,
	"CHEAP":          2,
	"NORMAL":         3,
	"EXPENSIVE":      4,
	"VERY_EXPENSIVE": 5,
}
