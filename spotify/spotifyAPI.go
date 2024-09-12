package spotify

import (
	"encoding/json"
	"fmt"
	"net/url"
	"github.com/spf13/viper"
    // "strings"
)


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

// func GetDataAPI (startDate, endDate, endpoint string) (map[string]interface{}, error) {
func GetDataAPI (startDate, endDate, endpoint string) ([]byte, error) {
	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)

	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/" + endpoint + "?" + params.Encode()

	// body := spotifyGETRequest(spotifyURL)
	//
	//    var data map[string]interface{}
	//    decoder := json.NewDecoder(strings.NewReader(body))
	//    if err := decoder.Decode(&data); err != nil {
	//        return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	//    }
	//    return data, nil

	body := spotifyGETRequest(spotifyURL)
    // body, err := spotifyGETRequest(spotifyURL)
    // if err != nil {
    //     return nil, fmt.Errorf("failed to get data from API: %w", err)
    // }

    return []byte(body), nil
}
