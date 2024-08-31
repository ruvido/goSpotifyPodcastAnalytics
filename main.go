package main

import (
	// "crypto/sha256"
	// "encoding/base64"
	// "encoding/json"
	"fmt"
	"io"
	"log"
	// "math/rand"
	"net/http"
	"net/url"
	"os"
	// "strings"
	"time"

	// "github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
    "github.com/ruvido/goSpotifyPodcastAnalytics/data"
	"github.com/ruvido/goSpotifyPodcastAnalytics/spotify"
	//"github.com/ruvido/goSpotifyPodcastAnalytics/caddy"
)

var (
	spotifyOAuthURL   = "https://accounts.spotify.com/oauth2/v2/auth"
	spotifyTokenURL   = "https://accounts.spotify.com/api/token"
	spotifyGenericURL = "https://generic.wg.spotify.com/podcasters/v0"
	filter            string
	lastDays          int
	outputJson		  string
)

func loadConfig() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error reading .env file: %v", err)
	}
	viper.AutomaticEnv()
}

func getDateRange() (startDate, endDate string) {
	if lastDays >= 0 {
		endDate = time.Now().Format("2006-01-02")
		startDate = time.Now().AddDate(0, 0, -lastDays).Format("2006-01-02")
	} else {
		// Default to all data
		startDate = "2020-01-01" // Assume this is the earliest data available
		endDate = time.Now().Format("2006-01-02")
	}
	return
}

func getSpotifyStreams(accessToken, startDate, endDate string) {
	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)
	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/detailedStreams" + "?" + params.Encode()

	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
}

func getSpotifyEpisodes(accessToken, startDate, endDate string) {
	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)
	// params.Set("page", "1")
	// params.Set("size", "100")
	// params.Set("sortBy", "releaseDate")
	// https://generic.wg.spotify.com/podcasters/v0/shows/0KkYBqKDT0iZVnUrpUcHS0/episodes?end=2024-08-27&filter=&page=1&size=50&sortBy=releaseDate&sortOrder=descending&start=2024-08-21
	
	// spotifyURL := spotifyGenericURL + "/shows/" + showID + "/episodes" + "?" + params.Encode()
	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/listeners" + "?" + params.Encode()
	
	
//	spotifyURL

	req, err := http.NewRequest("GET", spotifyURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))
}

// func SpotifyAnalytics(endpoint, startDate, endDate string) {
// 	showID := viper.GetString("SHOW_ID")
// 	params := url.Values{}
// 	params.Set("start", startDate)
// 	params.Set("end", endDate)
// 
// 	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/" + endpoint + "?" + params.Encode()
// 
// 	body := spotifyGETRequest(spotifyURL)
// 	
// 	fmt.Println(body)
// 	
// 	jsonKey := ""
// 	switch endpoint {
// 		case "detailedStreams":
// 			jsonKey = "detailedStreams"
// 			var result map[string][]lib.SpotifyDetailedStreams
// 	case "Listeners":
// 			jsonKey = "counts"
// 			var result map[string][]lib.SpotifyListeners
// 		default:
// 			log.Panic(endpoint,": this spotify endpoint is not yet supported :()")
// 	}
// 	
// 	// Unmarshal the JSON into the map
// 	err := json.Unmarshal([]byte(body), &result)
// 	if err != nil {
// 		log.Panic("Error unmarshaling JSON:", err)
// 	}
// 
// 	// Unmarshal JSON into the result map
// 	err := json.Unmarshal([]byte(body), &result)
// 	if err != nil {
// 		log.Panic("Error unmarshaling JSON:", err)
// 	}
// 	
// 	// Access the array of TimeAnalytics
// 	fmt.Println(result[jsonKey])
// 	return result[jsonKey]
// }


var rootCmd = &cobra.Command{
	Use:   "podcast-analytics",
	Short: "Podcast Analyitcs CLI tool",
}

var streamsCmd = &cobra.Command{
	Use:   "streams",
	Short: "Get Podcast Streams",
	Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("> STREAMS")
// 		startDate, endDate := getDateRange()
// 		filePath := viper.GetString("LOG_PATH")
// 		data := caddy.LoadLogData(filePath)
// 		filteredData := caddy.FilterLogData(data, startDate, endDate, filter)
// 		result := caddy.CountStreamsAndListeners(filteredData)
// 		err := caddy.OutputResult(result, outputJson)
// 		if err != nil {
// 			fmt.Printf("Error saving result: %v\n", err)
// 		}
// 
// 		//=======================================
// 		out := SpotifyAnalytics("detailedStreams", startDate, endDate)
// 		fmt.Println("spotify",out)
// 		fmt.Println("")
// 		//=======================================

	},
}


var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Podcast Episodes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing episodes...")
		// Implement listing logic here
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Podcast Analytics Summary",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Analytics summary...")
		// Implement summary logic here
	},
}

func init() {
	rootCmd.PersistentFlags().IntVar(&lastDays, "last", -1, "Number of last days to include (default: all data)")
	rootCmd.PersistentFlags().StringVar(&filter, "filter", "", "Filter episode names, number or season")
	rootCmd.PersistentFlags().StringVar(&outputJson, "json", "", "Output json filepath")

	rootCmd.AddCommand(streamsCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(listenersCmd)
	rootCmd.AddCommand(testCmd)

}

func main() {
	loadConfig()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


var listenersCmd = &cobra.Command{
	Use:   "listeners",
	Short: "Get Podcast Listeners",
	Run: func(cmd *cobra.Command, args []string) {
        //=======================================
		fmt.Println("> LISTENERS")
		startDate, endDate := getDateRange()
		endpoint := "listeners"
		listenrs, _ := spotify.Analytics(startDate, endDate, endpoint)
		// if err != nil {
		// 	log.Panic("Error in ingesting Spotify data")
		// }
        if slice, ok := listenrs.([]spotify.ListenersData); ok {
            for _, item := range slice {
                fmt.Printf("Date: %s, Count: %d\n", item.Date, item.Count)
            }
        } else {
            fmt.Println("Not a slice of ListenersData")
        }
        fmt.Println(listenrs)
    },
}

var testCmd= &cobra.Command{
    Use:   "test",
    Short: "Analytics test",
    Run: func(cmd *cobra.Command, args []string) {
        //=======================================
        fmt.Println("> TEST")
        startDate, endDate := getDateRange()
        endpoints := []string{"listeners", "detailedStreams"}

        sptfy, err := spotify.TimeAnalytics(startDate, endDate, endpoints)
        if err != nil {
            fmt.Println("Error (spotify):", err)
            return
        }

        // Initialize the original TimeAnalytics struct
        var original data.TimeAnalytics
        original.Name = sptfy

        // Print the final result
        fmt.Println(original)
        fmt.Println(original.Name["spotify"])
        //=======================================
    },
}
