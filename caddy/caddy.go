package caddy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"net/url"
	"strings"
)

type Request struct {
	URI     string              `json:"uri"`
	Headers map[string][]string `json:"headers"`
}

type LogEntry struct {
	Ts      float64 `json:"ts"`
	Request Request `json:"request"`
	Size    int64   `json:"size"`
}

type LogData struct {
	Timestamp string // The formatted date and time string
	RealIP    string // The real IP address extracted from headers
	URI       string // The URI from the log entry
	UserAgent string // The User-Agent string extracted from headers
	Size      int64  // The size of the log entry
}



type Result struct {
	TimeSeries []TimeSeries `json:"timeSeries"`
}

type TimeAnalytics struct {
	Date		string 		`json:"date"`
	Count		int 		`json:"count"`	
}


func containsAny(uri, filter string) bool {
	keywords := strings.Fields(filter)

	for _, keyword := range keywords {
		if !strings.Contains(uri, keyword) {
			return false
		}
	}
	return true
}


func ingestDataFromFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var logEntries []LogEntry
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read line: %w", err)
		}

		// Clean the line by decoding URL-encoded characters
		cleanLine, err := url.QueryUnescape(string(line))
		if err != nil {
			fmt.Printf("Failed to decode line: %v\n", err)
			continue
		}

		// Ensure the line has enough content to be a valid log entry
		if len(strings.TrimSpace(cleanLine)) == 0 || !strings.Contains(cleanLine, "}") {
			fmt.Println("Skipping incomplete or malformed log entry")
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(cleanLine), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}


		logEntries = append(logEntries, entry)
	}

	return logEntries, nil
}

func LoadLogData(filePath string) []LogData {
	logEntries, err := ingestDataFromFile(filePath)
	if err != nil {
		log.Fatalf("Failed to ingest data: %v", err)
	}

	var logDataList []LogData

	for _, entry := range logEntries {
		// Format the timestamp
		timestamp := time.Unix(int64(entry.Ts), 0).Format("2006-01-02 15:04:05")

		// Extract the real IP
		realIP := ""
		if ips, found := entry.Request.Headers["X-Real-Ip"]; found && len(ips) > 0 {
			realIP = ips[0]
		} else if ips, found := entry.Request.Headers["X-Forwarded-For"]; found && len(ips) > 0 {
			realIP = ips[0]
		}

		// Extract the User-Agent
		userAgent := ""
		if uas, found := entry.Request.Headers["User-Agent"]; found && len(uas) > 0 {
			userAgent = uas[0]
		}

		// Create an instance of LogData
		logData := LogData{
			Timestamp: timestamp,
			RealIP:    realIP,
			URI:       entry.Request.URI,
			UserAgent: userAgent,
			Size:      entry.Size,
		}

		// Add to the list
		if (entry.Size > 0 ) {
			logDataList = append(logDataList, logData)		
		}
	

	}
	return logDataList
}

func FilterLogData(data []LogData, startDate, endDate, filter string) []LogData {
	// Parse the start and end dates
	startDate = startDate + " 00:00:00"
	endDate = endDate + " 23:59:59"
	layout := "2006-01-02 15:04:05"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		fmt.Printf("Failed to parse start date: %v\n", err)
		return nil
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		fmt.Printf("Failed to parse end date: %v\n", err)
		return nil
	}

	var filteredData []LogData
	for _, entry := range data {
		entryTime, err := time.Parse(layout, entry.Timestamp)
		if err != nil {
			fmt.Printf("Failed to parse entry timestamp: %v\n", err)
			continue
		}

		if (entryTime.Equal(start) || entryTime.After(start)) && entryTime.Before(end) && containsAny(entry.URI, filter) {
			filteredData = append(filteredData, entry)
		}
	}

	return filteredData
}

func classifyUserAgent(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "spotify") {
		return "spotify"
	} else if strings.Contains(ua, "chrome") || strings.Contains(ua, "firefox") ||
		strings.Contains(ua, "safari") || strings.Contains(ua, "edge") ||
		strings.Contains(ua, "msie") || strings.Contains(ua, "opera") ||
		strings.Contains(ua, "mobile") {
		return "web"
	} else {
		return "other"
	}
}

func CountStreamsAndListeners(data []LogData) Result {
	streamsMap := make(map[string]map[string]struct{})
	listenersMap := make(map[string]map[string]struct{})
	dateCount := make(map[string]TimeSeries)

	for _, entry := range data {
		if entry.Size <= 0 {
			continue
		}

		date := entry.Timestamp[:10] // Extract the date (YYYY-MM-DD)
		category := classifyUserAgent(entry.UserAgent)

		epKey := entry.URI + entry.RealIP + entry.UserAgent
		listenerKey := entry.RealIP + entry.UserAgent

		if _, ok := streamsMap[date]; !ok {
			streamsMap[date] = make(map[string]struct{})
			listenersMap[date] = make(map[string]struct{})
		}

		if _, seen := streamsMap[date][epKey]; !seen {
			streamsMap[date][epKey] = struct{}{}
			dateCount[date] = incrementCount(dateCount[date], category, true)
			// if (category=="other") { fmt.Println(entry.UserAgent)}
		}

		if _, seen := listenersMap[date][listenerKey]; !seen {
			listenersMap[date][listenerKey] = struct{}{}
			dateCount[date] = incrementCount(dateCount[date], category, false)
		}
	}

	var result Result
	for _, v := range dateCount {
		result.TimeSeries = append(result.TimeSeries, v)
	}

	return result
}

func incrementCount(ts TimeSeries, category string, isStream bool) TimeSeries {
	if isStream {
		ts.All.Streams++
		switch category {
		case "web":
			ts.Web.Streams++
		case "spotify":
			ts.Spotify.Streams++
		case "other":
			ts.Other.Streams++
		}
	} else {
		ts.All.Listeners++
		switch category {
		case "web":
			ts.Web.Listeners++
		case "spotify":
			ts.Spotify.Listeners++
		case "other":
			ts.Other.Listeners++
		}
	}
	return ts
}

func OutputResult(result Result, outputFilePath string) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "")

	// Print to stdout
	err := encoder.Encode(result)
	if err != nil {
		return fmt.Errorf("failed to encode JSON to stdout: %v", err)
	}

	if outputFilePath != "" {
		// Save to the specified file
		file, err := os.Create(outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "")
		err = encoder.Encode(result)
		if err != nil {
			return fmt.Errorf("failed to encode JSON to file: %v", err)
		}
	}

	return nil
}

