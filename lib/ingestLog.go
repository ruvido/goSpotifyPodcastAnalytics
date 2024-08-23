package lib

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

func FilterLogData ( data []LogData, startDate, endDate, filter string) {
	// Parse the start and end dates
	startDate = startDate+" 00:00:00"
	endDate = endDate + " 23:59:00"
	layout := "2006-01-02 15:04:05"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		fmt.Printf("Failed to parse start date: %v\n", err)
		// return nil
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		fmt.Printf("Failed to parse end date: %v\n", err)
		// return nil
	}

	var filteredData []LogData

	for _, entry := range data {
		// Parse the entry's timestamp
		entryTime, err := time.Parse(layout, entry.Timestamp)
		if err != nil {
			fmt.Printf("Failed to parse entry timestamp: %v\n", err)
			continue
		}

		// Check if entry falls within the date range
		if (entryTime.Equal(start) || entryTime.After(start)) && entryTime.Before(end) && containsAny(entry.URI, filter) {
			filteredData = append(filteredData, entry)
			fmt.Println(entry.Timestamp,entry.URI, entry.Size)
		}
	}
	
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
