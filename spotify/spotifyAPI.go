package spotify

import (
	"encoding/json"
	"fmt"
	"net/url"
	"github.com/spf13/viper"
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

// Process and unmarshal data for 'listeners' endpoint
func processListenersData(body string) ([]ListenersData, error) {
	var countData struct {
		Counts []ListenersData `json:"counts"`
	}
	if err := json.Unmarshal([]byte(body), &countData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal counts data: %w", err)
	}

	return countData.Counts, nil
}

// Process and unmarshal data for 'detailedStreams' endpoint
func processDetailedStreamsData(body string) ([]DetailedStreamsData, error) {
	var detailedStreamData struct {
		DetailedStreams []DetailedStreamsData `json:"detailedStreams"`
	}
	if err := json.Unmarshal([]byte(body), &detailedStreamData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal detailedStreams data: %w", err)
	}

	return detailedStreamData.DetailedStreams, nil
}


func Analytics(startDate, endDate, endpoint string) (interface{}, error) {
	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)

	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/" + endpoint + "?" + params.Encode()

	body := spotifyGETRequest(spotifyURL)

	switch endpoint {
	case "listeners":
		return processListenersData(body)
	case "detailedStreams":
		return processDetailedStreamsData(body)
	default:
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}
