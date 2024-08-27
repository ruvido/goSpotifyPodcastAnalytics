package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
	"strings"
	"math/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
	"github.com/PuerkitoBio/goquery"


)

var (
	spotifyOAuthURL   = "https://accounts.spotify.com/oauth2/v2/auth"
	spotifyTokenURL   = "https://accounts.spotify.com/api/token"
	spotifyGenericURL = "https://generic.wg.spotify.com/podcasters/v0"
	filter            string
	lastDays          int
	outputJson		  string
)


// SPOTIFY json types ========================
var SpotifyDetailedStreams struct {
	SpotifyStreamsElement []struct {
		Date    string 	`json:"date"`
		Starts  int    	`json:"starts"`
		Streams	int    	`json:"streams"`
	} `json:"detailedStreams"`
}
var SpotifyListeners struct {
	SpotifyListenersElement []struct {
		Date    string 	`json:"date"`
		Count	int   	`json:"streams"`
	} `json:"counts"`
}
//============================================

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


func spotifyGETRequest(spotifyURL string) string {
	accessToken := GetSpotifyAccessToken()
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

	return string(body)
}


func SpotifyStreams(startDate, endDate string) {
	endpoint := "streams"

	showID := viper.GetString("SHOW_ID")
	params := url.Values{}
	params.Set("start", startDate)
	params.Set("end", endDate)

	spotifyURL := spotifyGenericURL + "/shows/" + showID + "/" + endpoint + "?" + params.Encode()

	body := spotifyGETRequest(spotifyURL)
	
	fmt.Println(body)
	
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
}
