package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/binance-live/internal/config"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WSClient handles WebSocket connections to Binance
type WSClient struct {
	baseURL              string
	conn                 *websocket.Conn
	mu                   sync.RWMutex
	logger               *zap.Logger
	reconnectDelay       time.Duration
	maxReconnectAttempts int
	pingInterval         time.Duration
	handlers             map[string]WSHandler
	stopChan             chan struct{}
	doneChan             chan struct{}
}

// WSHandler is a function that handles WebSocket messages
type WSHandler func(message []byte) error

// NewWSClient creates a new WebSocket client
func NewWSClient(cfg *config.BinanceConfig, streamCfg *config.StreamConfig, logger *zap.Logger) *WSClient {

	return &WSClient{
		baseURL:              cfg.WSURL,
		logger:               logger,
		reconnectDelay:       time.Duration(streamCfg.ReconnectDelay) * time.Second,
		maxReconnectAttempts: streamCfg.MaxReconnectAttempts,
		pingInterval:         time.Duration(streamCfg.PingInterval) * time.Second,
		handlers:             make(map[string]WSHandler),
		stopChan:             make(chan struct{}),
		doneChan:             make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection with streams
func (c *WSClient) Connect(ctx context.Context, streams []string) error {

	streamPath := strings.Join(streams, "/")
	url := fmt.Sprintf("%s/stream?streams=%s", c.baseURL, streamPath)

	c.logger.Info("Connecting to Binance WebSocket", zap.String("url", url))

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {

		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	c.logger.Info("WebSocket connection established")
	return nil
}

// Start starts the WebSocket client with automatic reconnection
func (c *WSClient) Start(ctx context.Context, streams []string) error {

	attempt := 0

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopChan:
			return nil
		default:
		}

		// Connect
		if err := c.Connect(ctx, streams); err != nil {

			attempt++
			if attempt >= c.maxReconnectAttempts {

				return fmt.Errorf("max reconnect attempts reached: %w", err)
			}

			c.logger.Warn("Failed to connect, retrying",
				zap.Error(err),
				zap.Int("attempt", attempt),
				zap.Duration("delay", c.reconnectDelay),
			)

			time.Sleep(c.reconnectDelay)
			continue
		}

		// Reset attempt counter on successful connection
		attempt = 0

		// Start ping/pong handler
		go c.pingHandler(ctx)

		// Start reading messages
		if err := c.readMessages(ctx); err != nil {

			c.logger.Error("WebSocket read error", zap.Error(err))
			c.closeConnection()

			// Wait before reconnecting
			select {
			case <-ctx.Done():

				return ctx.Err()
			case <-c.stopChan:

				return nil
			case <-time.After(c.reconnectDelay):

				c.logger.Info("Attempting to reconnect...")
			}
		}
	}
}

// readMessages reads and processes incoming WebSocket messages
func (c *WSClient) readMessages(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():

			return ctx.Err()
		case <-c.stopChan:

			return nil
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {

			return fmt.Errorf("connection is nil")
		}

		_, message, err := conn.ReadMessage()
		if err != nil {

			return fmt.Errorf("read message error: %w", err)
		}

		// Parse the stream message
		var streamMsg struct {
			Stream string          `json:"stream"`
			Data   json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(message, &streamMsg); err != nil {

			c.logger.Warn("Failed to unmarshal stream message",
				zap.Error(err),
				zap.String("message", string(message)),
			)
			continue
		}

		// Call the appropriate handler
		c.mu.RLock()
		handler, exists := c.handlers[streamMsg.Stream]
		c.mu.RUnlock()

		if exists {

			if err := handler(streamMsg.Data); err != nil {
				
				c.logger.Error("Handler error",
					zap.String("stream", streamMsg.Stream),
					zap.Error(err),
				)
			}
		}
	}
}

// pingHandler sends periodic ping messages to keep connection alive
func (c *WSClient) pingHandler(ctx context.Context) {

	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():

			return
		case <-c.stopChan:
			
			return
		case <-ticker.C:

			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn != nil {
				if err := conn.WriteControl(
					
					websocket.PingMessage,
					[]byte{},
					time.Now().Add(10*time.Second),
				); err != nil {

					c.logger.Warn("Failed to send ping", zap.Error(err))
				}
			}
		}
	}
}

// RegisterHandler registers a handler for a specific stream
func (c *WSClient) RegisterHandler(stream string, handler WSHandler) {
	
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[stream] = handler
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	
	close(c.stopChan)
	return c.closeConnection()
}

// closeConnection closes the underlying WebSocket connection
func (c *WSClient) closeConnection() error {
	
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {

		err := c.conn.WriteMessage(

			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {

			c.logger.Warn("Failed to send close message", zap.Error(err))
		}

		c.conn.Close()
		c.conn = nil
		c.logger.Info("WebSocket connection closed")
	}

	return nil
}

// BuildStreamNames builds WebSocket stream names for symbols
func BuildStreamNames(symbols []string, intervals []string) []string {

	var streams []string

	for _, symbol := range symbols {

		symbolLower := strings.ToLower(symbol)

		// Add kline streams for each interval
		for _, interval := range intervals {

			streams = append(streams, fmt.Sprintf("%s@kline_%s", symbolLower, interval))
		}

		// Add ticker stream
		streams = append(streams, fmt.Sprintf("%s@ticker", symbolLower))

		// Add depth stream (update speed 1000ms)
		streams = append(streams, fmt.Sprintf("%s@depth@1000ms", symbolLower))

		// Add aggregated trade stream
		streams = append(streams, fmt.Sprintf("%s@aggTrade", symbolLower))
	}

	return streams
}

// GetStreamName extracts stream name from full stream path
func GetStreamName(fullStream string) (symbol, streamType, interval string) {

	parts := strings.Split(fullStream, "@")
	if len(parts) < 2 {

		return
	}

	symbol = strings.ToUpper(parts[0])
	streamType = parts[1]

	// Extract interval for kline streams
	if strings.HasPrefix(streamType, "kline_") {
		
		interval = strings.TrimPrefix(streamType, "kline_")
		streamType = "kline"
	}

	return
}
