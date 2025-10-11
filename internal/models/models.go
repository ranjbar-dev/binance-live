package models

import "time"

// Symbol represents a trading pair
type Symbol struct {
	ID         int       `db:"id"`
	Symbol     string    `db:"symbol"`
	BaseAsset  string    `db:"base_asset"`
	QuoteAsset string    `db:"quote_asset"`
	Status     string    `db:"status"`
	IsActive   bool      `db:"is_active"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// Kline represents candlestick/kline data
type Kline struct {
	Symbol              string    `db:"symbol"`
	Interval            string    `db:"interval"`
	OpenTime            time.Time `db:"open_time"`
	CloseTime           time.Time `db:"close_time"`
	OpenPrice           float64   `db:"open_price"`
	HighPrice           float64   `db:"high_price"`
	LowPrice            float64   `db:"low_price"`
	ClosePrice          float64   `db:"close_price"`
	Volume              float64   `db:"volume"`
	QuoteVolume         float64   `db:"quote_volume"`
	TradesCount         int       `db:"trades_count"`
	TakerBuyVolume      float64   `db:"taker_buy_volume"`
	TakerBuyQuoteVolume float64   `db:"taker_buy_quote_volume"`
	CreatedAt           time.Time `db:"created_at"`
}

// Ticker represents 24hr ticker price data
type Ticker struct {
	Symbol                string    `db:"symbol"`
	Timestamp             time.Time `db:"timestamp"`
	Price                 float64   `db:"price"`
	BidPrice              *float64  `db:"bid_price"`
	BidQty                *float64  `db:"bid_qty"`
	AskPrice              *float64  `db:"ask_price"`
	AskQty                *float64  `db:"ask_qty"`
	Volume24h             *float64  `db:"volume_24h"`
	QuoteVolume24h        *float64  `db:"quote_volume_24h"`
	PriceChange24h        *float64  `db:"price_change_24h"`
	PriceChangePercent24h *float64  `db:"price_change_percent_24h"`
	High24h               *float64  `db:"high_24h"`
	Low24h                *float64  `db:"low_24h"`
	TradesCount24h        *int      `db:"trades_count_24h"`
	CreatedAt             time.Time `db:"created_at"`
}

// DepthSnapshot represents order book depth snapshot
type DepthSnapshot struct {
	ID           int64     `db:"id"`
	Symbol       string    `db:"symbol"`
	Timestamp    time.Time `db:"timestamp"`
	LastUpdateID int64     `db:"last_update_id"`
	Bids         string    `db:"bids"` // JSON array of [price, quantity]
	Asks         string    `db:"asks"` // JSON array of [price, quantity]
	CreatedAt    time.Time `db:"created_at"`
}

// Trade represents an aggregated trade
type Trade struct {
	ID            int64     `db:"id"`
	Symbol        string    `db:"symbol"`
	TradeID       int64     `db:"trade_id"`
	Timestamp     time.Time `db:"timestamp"`
	Price         float64   `db:"price"`
	Quantity      float64   `db:"quantity"`
	QuoteQuantity float64   `db:"quote_quantity"`
	IsBuyerMaker  bool      `db:"is_buyer_maker"`
	CreatedAt     time.Time `db:"created_at"`
}

// SyncStatus tracks the synchronization status for each symbol and data type
type SyncStatus struct {
	Symbol       string    `db:"symbol"`
	DataType     string    `db:"data_type"`
	Interval     *string   `db:"interval"`
	LastSyncTime time.Time `db:"last_sync_time"`
	LastDataTime time.Time `db:"last_data_time"`
	Status       string    `db:"status"`
	ErrorMessage *string   `db:"error_message"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// LiveData represents real-time data to be published to Redis
type LiveData struct {
	Type      string                 `json:"type"` // "kline", "ticker", "depth", "trade"
	Symbol    string                 `json:"symbol"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
