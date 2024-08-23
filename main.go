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
	"github.com/spf13/viper"
)

var (
	spotifyOAuthURL   = "https://accounts.spotify.com/oauth2/v2/auth"
	spotifyTokenURL   = "https://accounts.spotify.com/api/token"
	spotifyGenericURL = "https://generic.wg.spotify.com/podcasters/v0"
//"https://generic.wg.spotify.com/podcasters/v0/shows/0KkYBqKDT0iZVnUrpUcHS0/detailedStreams"
)

func loadConfig() {
	// Set Viper to read from .env file (if present)
	viper.SetConfigFile(".env")

	// Read configuration from the .env file
	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error reading .env file: %v", err)
	}

	// Set Viper to read environment variables
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

func getSpotifyStreams(accessToken, startDate, endDate string) {
	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)
	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/detailedStreams" + "?" + params.Encode()
	// urlWithParams := fmt.Sprintf("%s?%s", spotifyGenericURL, params.Encode())
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

func main() {
	// Load configuration from environment variables and .env file
	loadConfig()

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

	// Step 3: Use the access token to make the API request
	startDate := "2024-08-21"
	endDate := "2024-08-21"
	getSpotifyStreams(accessToken, startDate, endDate)
}
