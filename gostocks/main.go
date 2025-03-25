package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	av "github.com/masonJamesWheeler/alpha-vantage-go-wrapper"
	chart "github.com/wcharczuk/go-chart"
)

const apiKey = "api_key"

type CandleData struct {
	C []float64 `json:"c"` // Closing prices
	H []float64 `json:"h"` // High prices
	L []float64 `json:"l"` // Low prices
	O []float64 `json:"o"` // Open prices
	S string    `json:"s"` // Status
	T []int64   `json:"t"` // Timestamps
	V []int64   `json:"v"` // Volume
}

func main() {
	fmt.Println("Enter stock ticker:")
	var ticker string
	fmt.Scanln(&ticker)

	data, err := getStockData(ticker)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if data.S != "ok" {
		fmt.Println("Error: API returned status", data.S)
		return
	}

	if len(data.T) == 0 || len(data.C) == 0 {
		fmt.Println("Error: No data available for the ticker", ticker)
		return
	}

	err = createChart(ticker, data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Chart saved as %s_chart.png\n", ticker)
}

func getStockData(ticker string) (*CandleData, error) {
	client := av.NewClient(apiKey)
	tsData, err := client.GetDailyTimeSeries(ticker)
	if err != nil {
		return nil, err
	}

	// Get current time and 365 days ago
	now := time.Now()
	from := now.AddDate(0, 0, -365)

	candleData := &CandleData{
		S: "ok",
	}

	for _, entry := range tsData.Data {
		t, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			continue
		}
		if t.Before(from) {
			continue
		}
		timestamp := t.Unix()

		open, _ := strconv.ParseFloat(entry.Open, 64)
		high, _ := strconv.ParseFloat(entry.High, 64)
		low, _ := strconv.ParseFloat(entry.Low, 64)
		close, _ := strconv.ParseFloat(entry.Close, 64)
		volume, _ := strconv.ParseInt(entry.Volume, 10, 64)

		candleData.T = append(candleData.T, timestamp)
		candleData.O = append(candleData.O, open)
		candleData.H = append(candleData.H, high)
		candleData.L = append(candleData.L, low)
		candleData.C = append(candleData.C, close)
		candleData.V = append(candleData.V, volume)
	}

	// Sort by timestamp in ascending order
	sort.Slice(candleData.T, func(i, j int) bool {
		return candleData.T[i] < candleData.T[j]
	})

	return candleData, nil
}

func createChart(ticker string, data *CandleData) error {
	// Convert Unix timestamps to time.Time for the x-axis
	var times []time.Time
	for _, ts := range data.T {
		times = append(times, time.Unix(ts, 0))
	}

	// Create a time series for the closing prices
	series := chart.TimeSeries{
		Name:    "Closing Price",
		XValues: times,
		YValues: data.C,
	}

	// Configure the chart
	graph := chart.Chart{
		Title: fmt.Sprintf("%s Stock Price", ticker),
		XAxis: chart.XAxis{
			Name: "Date",
		},
		YAxis: chart.YAxis{
			Name: "Price",
		},
		Series: []chart.Series{series},
	}

	// Save the chart to a PNG file
	f, err := os.Create(fmt.Sprintf("%s_chart.png", ticker))
	if err != nil {
		return err
	}
	defer f.Close()

	err = graph.Render(chart.PNG, f)
	if err != nil {
		return err
	}

	return nil
}
