package spotify

import (
	// "encoding/json"
	// "fmt"
	// "net/url"
	// "github.com/spf13/viper"
    "github.com/ruvido/goSpotifyPodcastAnalytics/data"
)

// func TimeAnalytics(startDate, endDate string, endpoints []string) ([]data.DailyAnalytics, error) {
//
//     // Create and populate the slice of DailyAnalytics
//     dataMap := []data.DailyAnalytics{
//         {Date: "2024-08-30", Streams: 1000, Listeners: 800},
//         {Date: "2024-08-31", Streams: 1200, Listeners: 950},
//     }
//
//     // Return the map and no error
//     return dataMap, nil
// }
func TimeAnalytics(startDate, endDate string, endpoints []string) (map[string][]data.DailyAnalytics, error) {
    // Example data
    dataMap := map[string][]data.DailyAnalytics{
        "spotify": {
            {Date: "2024-08-30", Streams: 1000, Listeners: 800},
            {Date: "2024-08-31", Streams: 1200, Listeners: 950},
        },
        // Additional platforms can be added here
    }

    return dataMap, nil
}
