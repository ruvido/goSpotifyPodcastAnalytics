package spotify

import (
	"encoding/json"
	"fmt"
	// "net/url"
	// "github.com/spf13/viper"
    "github.com/ruvido/goSpotifyPodcastAnalytics/data"
    // "strings"
    // "errors"
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

// Aggregate daily stats
// type DailyAnalytics struct {
// 	Date     string
// 	Streams  int
// 	Listeners int
// }

func TimeAnalyticsBUM(startDate, endDate string, endpoints []string) (map[string][]data.DailyAnalytics, error) {
    dataMap := map[string][]data.DailyAnalytics{"spotify": {}}
    // var streams   []DetailedStreamsData
    // var listeners []ListenersData

    var streamsData []DetailedStreamsData
    var listenersData []ListenersData

    // Fetch and unmarshal data for each endpoint
    for _, endpoint := range endpoints {
        data, err := GetDataAPI(startDate, endDate, endpoint)
        if err != nil {
            return nil, err
        }

        switch endpoint {
        case "detailedStreams":
            if err := json.Unmarshal(data, &streamsData); err != nil {
                return nil, fmt.Errorf("failed to unmarshal detailedStreams data: %w", err)
            }
        case "listeners":
            if err := json.Unmarshal(data, &listenersData); err != nil {
                return nil, fmt.Errorf("failed to unmarshal counts data: %w", err)
            }
        default:
            return nil, fmt.Errorf("ERR Spotify: %s Unknown endpoint", endpoint)
        }
    }

    return dataMap, nil
}

        // switch endpoint {
        // case "detailedStreams":
        //     streams = data["detailedStreams"]
        // case "listeners":
        //     listeners = data["count"]
        // default:
        //     fmt.Printf("ERR Spotify: %s Unknown endpoint\n", endpoint)
        // }
    // }

    // if len(listeners) != len(streams) {
    //     return nil, errors.New("listeners and streams have different lengths")
    // }
    //
    // for i := range listeners {
    //     dataMap["spotify"] = append(dataMap["spotify"], data.DailyAnalytics{
    //         Date:      listeners[i].Date,
    //         Streams:   streams[i].Streams,
    //         Listeners: listeners[i].Count,
    //     })
    // }

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

