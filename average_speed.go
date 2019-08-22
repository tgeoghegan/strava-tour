package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/strava/go.strava"
)

func prettyJson(v interface{}) []byte {
	pretty, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return []byte(fmt.Sprintf("json marshal failure: %s", err))
	}
	return pretty
}

func main() {
	var athleteAccessToken string
	var appClientId int
	var appClientSecret string

	flag.StringVar(&athleteAccessToken, "token", "", "Access token")
	flag.IntVar(&appClientId, "client-id", 38247, "Application client ID")
	flag.StringVar(&appClientSecret, "client-secret", "", "Application client secret")

	flag.Parse()

	if athleteAccessToken == "" {
		if appClientId == 0 || appClientSecret == "" {
			fmt.Printf("Must provide either athlete access token or app client ID and app client secret\n")
			os.Exit(1)
		}

		var err error
		athleteAccessToken, err = Authorize(appClientId, appClientSecret)
		if err != nil {
			fmt.Printf("Failed to obtain athlete access token: %s\n", err)
			os.Exit(1)
		}
	}

	if athleteAccessToken == "" {
		fmt.Printf("No athlete access token\n")
		os.Exit(1)
	}

	// This call, and subsequent calls to create services, cannot fail as they
	// merely allocate objects.
	client := strava.NewClient(athleteAccessToken)

	// First, list the activities for the athlete implied by our access token
	athleteService := strava.NewCurrentAthleteService(client)

	// List activities between June 1, 2019 and August 17, 2019 in UTC,
	// hopefully corresponding to my tour this summer.
	start := time.Date(2019, time.June, 1, 0, 0, 0, 0, time.UTC).Unix()
	end := time.Date(2019, time.August, 17, 0, 0, 0, 0, time.UTC).Unix()
	activities, err := athleteService.ListActivities().
		After(int(start)).
		Before(int(end)).
		PerPage(200).
		Do()
	if err != nil {
		fmt.Printf("Failed to obtain activities list: %s\n", err)
		os.Exit(1)
	}

	// Take average over average speeds
	speedSum := 0.0
	for _, activity := range activities {
		speedSum += activity.AverageSpeed
	}

	// API provides speed in m/s, so convert to kph
	averageSpeedKilometersPerHour := speedSum / float64(len(activities)) / 1000 * 3600
	fmt.Printf("Average speed: %f kph\n", averageSpeedKilometersPerHour)
}
