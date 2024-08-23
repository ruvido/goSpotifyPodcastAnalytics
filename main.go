package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ruvido/goSpotifyPodcastAnalytics/lib"
)

var (
	spotifyOAuthURL   = "https://accounts.spotify.com/oauth2/v2/auth"
	spotifyTokenURL   = "https://accounts.spotify.com/api/token"
	spotifyGenericURL = "https://generic.wg.spotify.com/podcasters/v0"
	filter            string
	lastDays          int
)

func loadConfig() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error reading .env file: %v", err)
	}
	viper.AutomaticEnv()
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateCodeChallenge(codeVerifier string) string {
	sha256Hasher := sha256.New()
	sha256Hasher.Write([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(sha256Hasher.Sum(nil))
}

func getAuthorizationCode(codeChallenge, codeVerifier string) string {
	clientID := viper.GetString("CLIENT_ID")
	spDc := viper.GetString("SP_DC")
	spKey := viper.GetString("SP_KEY")
	redirectURI := "https://podcasters.spotify.com"

	baseURL := spotifyOAuthURL
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", clientID)
	params.Add("scope", "streaming ugc-image-upload user-read-email user-read-private")
	params.Add("redirect_uri", redirectURI)
	params.Add("code_challenge", codeChallenge)
	params.Add("code_challenge_method", "S256")
	params.Add("state", codeVerifier)
	params.Add("response_mode", "web_message")
	params.Add("prompt", "none")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Cookie", fmt.Sprintf("sp_dc=%s; sp_key=%s", spDc, spKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jsContent string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		jsContent = s.Text()
	})

	codeStart := strings.Index(jsContent, `"code": "`) + len(`"code": "`)
	codeEnd := strings.Index(jsContent[codeStart:], `"`)
	code := jsContent[codeStart : codeStart+codeEnd]

	return code
}

func getAccessToken(code, codeVerifier string) string {
	clientID := viper.GetString("CLIENT_ID")
	redirectURI := "https://podcasters.spotify.com"

	tokenURL := spotifyTokenURL
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Fatal(err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		log.Fatal("Failed to get access token from response")
	}

	return accessToken
}

func GetSpotifyAccessToken () string {
		// Load configuration from environment variables and .env file
	// loadConfig()

	// Step 1: Generate Code Verifier and Code Challenge
	codeVerifier := generateRandomString(64)
	codeChallenge := generateCodeChallenge(codeVerifier)
	code := getAuthorizationCode(codeChallenge, codeVerifier)
	if code == "" {
		log.Fatal("Failed to get authorization code")
	}

	// Step 2: Exchange the authorization code for an access token
	accessToken := getAccessToken(code, codeVerifier)
	if accessToken == "" {
		log.Fatal("Failed to get access token")
	}
	return accessToken
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

var rootCmd = &cobra.Command{
	Use:   "podcast-analytics",
	Short: "Podcast Analyitcs CLI tool",
}

var streamsCmd = &cobra.Command{
	Use:   "streams",
	Short: "Get Podcast Streams",
	Run: func(cmd *cobra.Command, args []string) {
		// accessToken := GetSpotifyAccessToken()
		startDate, endDate := getDateRange() // Get the date range based on the global flag
		// getSpotifyStreams(accessToken, startDate, endDate)
		filePath := viper.GetString("LOG_PATH")
	
		data := lib.LoadLogData(filePath)
		lib.FilterLogData(data, startDate, endDate, filter)
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

	rootCmd.AddCommand(streamsCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(summaryCmd)
}

func main() {
	loadConfig()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
