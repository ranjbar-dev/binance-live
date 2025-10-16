# PowerShell script to benchmark protobuf vs JSON performance
# This script demonstrates the performance benefits of protobuf

Write-Host "Protobuf vs JSON Performance Benchmark"
Write-Host "====================================="

# Create a simple Go benchmark program
$benchmarkCode = @'
package main

import (
	"encoding/json"
	"fmt"
	"time"

	binanceProto "github.com/binance-live/proto"
	"google.golang.org/protobuf/proto"
)

type JSONKlineData struct {
	Symbol              string  `json:"symbol"`
	Interval            string  `json:"interval"`
	OpenTime            int64   `json:"open_time"`
	CloseTime           int64   `json:"close_time"`
	OpenPrice           float64 `json:"open_price"`
	HighPrice           float64 `json:"high_price"`
	LowPrice            float64 `json:"low_price"`
	ClosePrice          float64 `json:"close_price"`
	Volume              float64 `json:"volume"`
	QuoteVolume         float64 `json:"quote_volume"`
	TradesCount         int32   `json:"trades_count"`
	TakerBuyVolume      float64 `json:"taker_buy_volume"`
	TakerBuyQuoteVolume float64 `json:"taker_buy_quote_volume"`
}

func main() {
	// Sample data
	symbol := "BTCUSDT"
	interval := "1m"
	openTime := int64(1640995200) // 2022-01-01 00:00:00 UTC
	closeTime := int64(1640995260) // 2022-01-01 00:01:00 UTC
	openPrice := 47000.0
	highPrice := 47500.0
	lowPrice := 46500.0
	closePrice := 47200.0
	volume := 100.5
	quoteVolume := 4720000.0
	tradesCount := int32(150)
	takerBuyVolume := 60.2
	takerBuyQuoteVolume := 2832000.0

	// Create protobuf data
	protoKline := &binanceProto.KlineData{
		Interval:              interval,
		OpenTime:              openTime,
		CloseTime:             closeTime,
		OpenPrice:             openPrice,
		HighPrice:             highPrice,
		LowPrice:              lowPrice,
		ClosePrice:            closePrice,
		Volume:                volume,
		QuoteVolume:           quoteVolume,
		TradesCount:           tradesCount,
		TakerBuyVolume:        takerBuyVolume,
		TakerBuyQuoteVolume:   takerBuyQuoteVolume,
	}

	protoLiveData := &binanceProto.LiveData{
		Type:      binanceProto.DataType_DATA_TYPE_KLINE,
		Symbol:    symbol,
		Timestamp: openTime * 1000, // Convert to milliseconds
		Data: &binanceProto.LiveData_Kline{
			Kline: protoKline,
		},
	}

	// Create JSON data
	jsonKline := JSONKlineData{
		Symbol:              symbol,
		Interval:            interval,
		OpenTime:            openTime,
		CloseTime:           closeTime,
		OpenPrice:           openPrice,
		HighPrice:           highPrice,
		LowPrice:            lowPrice,
		ClosePrice:          closePrice,
		Volume:              volume,
		QuoteVolume:         quoteVolume,
		TradesCount:         tradesCount,
		TakerBuyVolume:      takerBuyVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
	}

	jsonLiveData := map[string]interface{}{
		"type":      "kline",
		"symbol":    symbol,
		"timestamp": openTime * 1000,
		"data":      jsonKline,
	}

	// Benchmark protobuf marshaling
	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := proto.Marshal(protoLiveData)
		if err != nil {
			panic(err)
		}
	}
	protoMarshalTime := time.Since(start)

	// Benchmark protobuf unmarshaling
	protoData, _ := proto.Marshal(protoLiveData)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var result binanceProto.LiveData
		err := proto.Unmarshal(protoData, &result)
		if err != nil {
			panic(err)
		}
	}
	protoUnmarshalTime := time.Since(start)

	// Benchmark JSON marshaling
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := json.Marshal(jsonLiveData)
		if err != nil {
			panic(err)
		}
	}
	jsonMarshalTime := time.Since(start)

	// Benchmark JSON unmarshaling
	jsonData, _ := json.Marshal(jsonLiveData)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var result map[string]interface{}
		err := json.Unmarshal(jsonData, &result)
		if err != nil {
			panic(err)
		}
	}
	jsonUnmarshalTime := time.Since(start)

	// Calculate sizes
	protoSize := len(protoData)
	jsonSize := len(jsonData)

	// Print results
	fmt.Printf("Benchmark Results (%d iterations):\n", iterations)
	fmt.Printf("=====================================\n")
	fmt.Printf("Protobuf Marshal Time:   %v\n", protoMarshalTime)
	fmt.Printf("Protobuf Unmarshal Time: %v\n", protoUnmarshalTime)
	fmt.Printf("JSON Marshal Time:       %v\n", jsonMarshalTime)
	fmt.Printf("JSON Unmarshal Time:     %v\n", jsonUnmarshalTime)
	fmt.Printf("\n")
	fmt.Printf("Protobuf Size: %d bytes\n", protoSize)
	fmt.Printf("JSON Size:     %d bytes\n", jsonSize)
	fmt.Printf("Size Reduction: %.1f%%\n", float64(jsonSize-protoSize)/float64(jsonSize)*100)
	fmt.Printf("\n")
	fmt.Printf("Protobuf Marshal Speedup:   %.1fx\n", float64(jsonMarshalTime)/float64(protoMarshalTime))
	fmt.Printf("Protobuf Unmarshal Speedup: %.1fx\n", float64(jsonUnmarshalTime)/float64(protoUnmarshalTime))
}
'@

# Write benchmark code to file
$benchmarkCode | Out-File -FilePath "benchmark.go" -Encoding UTF8

Write-Host "Running benchmark..."
go run benchmark.go

# Clean up
Remove-Item "benchmark.go" -Force

Write-Host ""
Write-Host "Benchmark completed!"
