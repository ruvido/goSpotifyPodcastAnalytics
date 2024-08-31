package data

import (
    // "fmt"
)

type DailyAnalytics struct {
    Date        string  `json:"date"`
    Streams     int     `json:"streams"`
    Listeners   int     `json:"listeners"`
}

type TimeAnalytics struct {
    Name map[string][]DailyAnalytics `json:"platforms"`
}
