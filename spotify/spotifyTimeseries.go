package spotify

import (
	"encoding/json"
	"fmt"
	// "net/url"
	// "github.com/spf13/viper"
    // "strings"
    // "errors"
    "github.com/ruvido/goSpotifyPodcastAnalytics/data"
)

// Data structure for listeners endpoint
type ListenersData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// Data structure for detailedStreams endpoint
type DetailedStreamsData struct {
	Date    string `json:"date"`
	Starts  int    `json:"starts"`
	Streams int    `json:"streams"`
}

// Define a struct to match the API response format
type ResponseData struct {
    Counts           []ListenersData      `json:"counts"`
    DetailedStreams  []DetailedStreamsData `json:"detailedStreams"`
}

func TimeAnalytics(startDate, endDate string, endpoints []string) (map[string][]data.DailyAnalytics, error) {
    var allData ResponseData

    // Fetch and unmarshal data for each endpoint
    for _, endpoint := range endpoints {
        jsonData, err := GetDataAPI(startDate, endDate, endpoint)
        if err != nil {
            return nil, err
        }

        // Unmarshal JSON byte slice into ResponseData struct
        if err := json.Unmarshal(jsonData, &allData); err != nil {
            return nil, fmt.Errorf("failed to unmarshal response data: %w", err)
        }
    }

    // Create a map to combine the data
    dataMap := make(map[string][]data.DailyAnalytics)
    dataMap["spotify"] = []data.DailyAnalytics{}

    // Create a map to store streams by date
    streamsMap := make(map[string]DetailedStreamsData)
    for _, stream := range allData.DetailedStreams {
        streamsMap[stream.Date] = stream
    }

    // Combine data
    for _, listener := range allData.Counts {
        stream, exists := streamsMap[listener.Date]
        if !exists {
            return nil, fmt.Errorf("no streams data for date: %s", listener.Date)
        }
        dataMap["spotify"] = append(dataMap["spotify"], data.DailyAnalytics{
            Date:      listener.Date,
            Streams:   stream.Streams,
            Listeners: listener.Count,
        })
    }

    return dataMap, nil
}

