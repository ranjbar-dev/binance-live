package binance

import (
	"github.com/binance-live/internal/config"
	"go.uber.org/zap"
)

// Client is the main Binance API client that wraps both REST and WebSocket clients
type Client struct {
	REST      *RESTClient
	WebSocket *WSClient
	Config    *config.BinanceConfig
	Logger    *zap.Logger
}

// NewClient creates a new Binance API client
func NewClient(cfg *config.Config, logger *zap.Logger) *Client {
	return &Client{
		REST:      NewRESTClient(&cfg.Binance, logger),
		WebSocket: NewWSClient(&cfg.Binance, &cfg.Stream, logger),
		Config:    &cfg.Binance,
		Logger:    logger,
	}
}
