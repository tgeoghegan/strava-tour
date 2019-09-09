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

	// We have to get an activity service to go from activity summaries to
	// detailed activity information, so that we can then get information on
	// segment efforts and finally the segments -- and hill categories.
	activityService := strava.NewActivitiesService(client)

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
	distanceSum := 0.0
	elevationSum := 0.0
	categories := map[strava.ClimbCategory]int{
		strava.ClimbCategories.NotCategorized: 0,
		strava.ClimbCategories.Category1:      0,
		strava.ClimbCategories.Category2:      0,
		strava.ClimbCategories.Category3:      0,
		strava.ClimbCategories.Category4:      0,
		strava.ClimbCategories.HorsCategorie:  0,
	}
	var biggestClimb *strava.SegmentSummary
	var toughestAverageGrade *strava.SegmentSummary
	var toughestMaximumGrade *strava.SegmentSummary

	for _, activity := range activities {
		speedSum += activity.AverageSpeed
		distanceSum += activity.Distance
		elevationSum += activity.TotalElevationGain

		activityDetail, err := activityService.Get(activity.Id).
			IncludeAllEfforts().
			Do()
		if err != nil {
			fmt.Printf("Failed to obtain detailed activity %d: %s\n", activity.Id, err)
			os.Exit(1)
		}

		for _, segmentEffortSummary := range activityDetail.SegmentEfforts {
			categories[segmentEffortSummary.Segment.ClimbCategory] += 1

			segmentElevation := segmentEffortSummary.Segment.ElevationHigh - segmentEffortSummary.Segment.ElevationLow
			// Check average grade to filter out downhill segments
			if (biggestClimb == nil || segmentElevation > (biggestClimb.ElevationHigh-biggestClimb.ElevationLow)) &&
				segmentEffortSummary.Segment.AverageGrade > 0 {
				biggestClimb = &segmentEffortSummary.Segment
			}
			if toughestAverageGrade == nil || segmentEffortSummary.Segment.AverageGrade > toughestAverageGrade.AverageGrade {
				toughestAverageGrade = &segmentEffortSummary.Segment
			}
			if toughestMaximumGrade == nil || segmentEffortSummary.Segment.MaximumGrade > toughestMaximumGrade.AverageGrade {
				toughestMaximumGrade = &segmentEffortSummary.Segment
			}
		}
	}

	numActivities := float64(len(activities))

	// API provides speed in m/s, so convert to kph
	averageSpeedKilometersPerHour := speedSum / numActivities / 1000 * 3600
	averageDistanceKilometers := distanceSum / numActivities / 1000
	averageElevation := elevationSum / numActivities
	fmt.Printf("Average speed: %f kph\nTotal distance: %f km\nTotal elevation: %f m\n"+
		"Average distance per day: %f km\nAverage elevation per day: %f m\n",
		averageSpeedKilometersPerHour, distanceSum/1000, elevationSum,
		averageDistanceKilometers, averageElevation)

	fmt.Printf("Hill categories:\n\tNot categorized: %d\n\tCategory 1: %d\n\tCategory 2: %d\n\tCategory 3: %d\n\tCategory 4: %d\n\tHors catégorie: %d\n",
		categories[strava.ClimbCategories.NotCategorized],
		categories[strava.ClimbCategories.Category1],
		categories[strava.ClimbCategories.Category2],
		categories[strava.ClimbCategories.Category3],
		categories[strava.ClimbCategories.Category4],
		categories[strava.ClimbCategories.HorsCategorie],
	)
	fmt.Printf("Biggest climb: %+v\n", biggestClimb)
	fmt.Printf("Toughest average grade: %+v\n", toughestAverageGrade)
	fmt.Printf("Toughest maximum grade: %+v\n", toughestMaximumGrade)
}
