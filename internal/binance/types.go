package binance

// KlineResponse represents a kline/candlestick from Binance API
type KlineResponse []interface{}

// ParseKlineResponse parses a kline response into structured data
func ParseKlineResponse(data KlineResponse) (*KlineData, error) {
	
	return &KlineData{
		OpenTime:                 int64(data[0].(float64)),
		Open:                     data[1].(string),
		High:                     data[2].(string),
		Low:                      data[3].(string),
		Close:                    data[4].(string),
		Volume:                   data[5].(string),
		CloseTime:                int64(data[6].(float64)),
		QuoteAssetVolume:         data[7].(string),
		NumberOfTrades:           int(data[8].(float64)),
		TakerBuyBaseAssetVolume:  data[9].(string),
		TakerBuyQuoteAssetVolume: data[10].(string),
	}, nil
}

// KlineData represents parsed kline data
type KlineData struct {
	OpenTime                 int64
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                int64
	QuoteAssetVolume         string
	NumberOfTrades           int
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
}

// Ticker24hrResponse represents 24hr ticker statistics
type Ticker24hrResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int    `json:"count"`
}

// DepthResponse represents order book depth
type DepthResponse struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"` // [price, quantity]
	Asks         [][]string `json:"asks"` // [price, quantity]
}

// AggTradeResponse represents aggregated trade data
type AggTradeResponse struct {
	AggTradeID   int64  `json:"a"` // Aggregate tradeId
	Price        string `json:"p"` // Price
	Quantity     string `json:"q"` // Quantity
	FirstTradeID int64  `json:"f"` // First tradeId
	LastTradeID  int64  `json:"l"` // Last tradeId
	Timestamp    int64  `json:"T"` // Timestamp
	IsBuyerMaker bool   `json:"m"` // Was the buyer the maker?
	IsBestMatch  bool   `json:"M"` // Was the trade the best price match?
}

// ExchangeInfoResponse represents exchange information
type ExchangeInfoResponse struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	Symbols    []SymbolInfo `json:"symbols"`
}

// SymbolInfo represents trading pair information
type SymbolInfo struct {
	Symbol     string `json:"symbol"`
	Status     string `json:"status"`
	BaseAsset  string `json:"baseAsset"`
	QuoteAsset string `json:"quoteAsset"`
}

// WebSocket Stream Messages

// WSKlineEvent represents a kline WebSocket event
type WSKlineEvent struct {
	EventType string  `json:"e"` // Event type
	EventTime int64   `json:"E"` // Event time
	Symbol    string  `json:"s"` // Symbol
	Kline     WSKline `json:"k"` // Kline data
}

// WSKline represents kline data in WebSocket event
type WSKline struct {
	StartTime                int64  `json:"t"` // Kline start time
	EndTime                  int64  `json:"T"` // Kline close time
	Symbol                   string `json:"s"` // Symbol
	Interval                 string `json:"i"` // Interval
	FirstTradeID             int64  `json:"f"` // First trade ID
	LastTradeID              int64  `json:"L"` // Last trade ID
	Open                     string `json:"o"` // Open price
	Close                    string `json:"c"` // Close price
	High                     string `json:"h"` // High price
	Low                      string `json:"l"` // Low price
	Volume                   string `json:"v"` // Base asset volume
	NumberOfTrades           int    `json:"n"` // Number of trades
	IsClosed                 bool   `json:"x"` // Is this kline closed?
	QuoteVolume              string `json:"q"` // Quote asset volume
	TakerBuyBaseAssetVolume  string `json:"V"` // Taker buy base asset volume
	TakerBuyQuoteAssetVolume string `json:"Q"` // Taker buy quote asset volume
}

// WSTickerEvent represents a 24hr ticker WebSocket event
type WSTickerEvent struct {
	EventType          string `json:"e"` // Event type
	EventTime          int64  `json:"E"` // Event time
	Symbol             string `json:"s"` // Symbol
	PriceChange        string `json:"p"` // Price change
	PriceChangePercent string `json:"P"` // Price change percent
	WeightedAvgPrice   string `json:"w"` // Weighted average price
	FirstPrice         string `json:"x"` // First trade(F)-1 price (first trade before the 24hr rolling window)
	LastPrice          string `json:"c"` // Last price
	LastQty            string `json:"Q"` // Last quantity
	BidPrice           string `json:"b"` // Best bid price
	BidQty             string `json:"B"` // Best bid quantity
	AskPrice           string `json:"a"` // Best ask price
	AskQty             string `json:"A"` // Best ask quantity
	OpenPrice          string `json:"o"` // Open price
	HighPrice          string `json:"h"` // High price
	LowPrice           string `json:"l"` // Low price
	Volume             string `json:"v"` // Total traded base asset volume
	QuoteVolume        string `json:"q"` // Total traded quote asset volume
	OpenTime           int64  `json:"O"` // Statistics open time
	CloseTime          int64  `json:"C"` // Statistics close time
	FirstID            int64  `json:"F"` // First trade ID
	LastID             int64  `json:"L"` // Last trade Id
	Count              int    `json:"n"` // Total number of trades
}

// WSDepthEvent represents a depth update WebSocket event
type WSDepthEvent struct {
	EventType     string     `json:"e"` // Event type
	EventTime     int64      `json:"E"` // Event time
	Symbol        string     `json:"s"` // Symbol
	FirstUpdateID int64      `json:"U"` // First update ID in event
	FinalUpdateID int64      `json:"u"` // Final update ID in event
	Bids          [][]string `json:"b"` // Bids to be updated [price, quantity]
	Asks          [][]string `json:"a"` // Asks to be updated [price, quantity]
}

// WSAggTradeEvent represents an aggregated trade WebSocket event
type WSAggTradeEvent struct {
	EventType    string `json:"e"` // Event type
	EventTime    int64  `json:"E"` // Event time
	Symbol       string `json:"s"` // Symbol
	AggTradeID   int64  `json:"a"` // Aggregate trade ID
	Price        string `json:"p"` // Price
	Quantity     string `json:"q"` // Quantity
	FirstTradeID int64  `json:"f"` // First trade ID
	LastTradeID  int64  `json:"l"` // Last trade ID
	TradeTime    int64  `json:"T"` // Trade time
	IsBuyerMaker bool   `json:"m"` // Is the buyer the market maker?
}

// Error Response

// APIError represents a Binance API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e *APIError) Error() string {
	return e.Message
}
