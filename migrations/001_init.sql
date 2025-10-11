-- Initial schema for Binance Live Data Collector
-- Using TimescaleDB for time-series data optimization

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table to store trading pair symbols
CREATE TABLE IF NOT EXISTS symbols (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    base_asset VARCHAR(10) NOT NULL,
    quote_asset VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'TRADING',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_symbols_active ON symbols(is_active);
CREATE INDEX idx_symbols_symbol ON symbols(symbol);

-- Table to store kline/candlestick data
CREATE TABLE IF NOT EXISTS klines (
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(5) NOT NULL,
    open_time TIMESTAMP WITH TIME ZONE NOT NULL,
    close_time TIMESTAMP WITH TIME ZONE NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    quote_volume DECIMAL(20, 8) NOT NULL,
    trades_count INTEGER NOT NULL,
    taker_buy_volume DECIMAL(20, 8) NOT NULL,
    taker_buy_quote_volume DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (symbol, interval, open_time)
);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('klines', 'open_time', if_not_exists => TRUE);

-- Create indexes for efficient queries
CREATE INDEX idx_klines_symbol_interval ON klines(symbol, interval, open_time DESC);

-- Table to store ticker/price data
CREATE TABLE IF NOT EXISTS tickers (
    symbol VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    bid_price DECIMAL(20, 8),
    bid_qty DECIMAL(20, 8),
    ask_price DECIMAL(20, 8),
    ask_qty DECIMAL(20, 8),
    volume_24h DECIMAL(20, 8),
    quote_volume_24h DECIMAL(20, 8),
    price_change_24h DECIMAL(20, 8),
    price_change_percent_24h DECIMAL(10, 4),
    high_24h DECIMAL(20, 8),
    low_24h DECIMAL(20, 8),
    trades_count_24h INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (symbol, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('tickers', 'timestamp', if_not_exists => TRUE);

CREATE INDEX idx_tickers_symbol ON tickers(symbol, timestamp DESC);

-- Table to store order book depth data
CREATE TABLE IF NOT EXISTS depth_snapshots (
    id BIGSERIAL,
    symbol VARCHAR(20) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    last_update_id BIGINT NOT NULL,
    bids JSONB NOT NULL, -- Array of [price, quantity]
    asks JSONB NOT NULL, -- Array of [price, quantity]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('depth_snapshots', 'timestamp', if_not_exists => TRUE);

CREATE INDEX idx_depth_symbol ON depth_snapshots(symbol, timestamp DESC);

-- Table to store aggregated trade data (quotes)
CREATE TABLE IF NOT EXISTS trades (
    id BIGSERIAL,
    symbol VARCHAR(20) NOT NULL,
    trade_id BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    quote_quantity DECIMAL(20, 8) NOT NULL,
    is_buyer_maker BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, timestamp),
    UNIQUE (symbol, trade_id)
);

-- Convert to hypertable
SELECT create_hypertable('trades', 'timestamp', if_not_exists => TRUE);

CREATE INDEX idx_trades_symbol ON trades(symbol, timestamp DESC);

-- Table to track sync status for each symbol
CREATE TABLE IF NOT EXISTS sync_status (
    symbol VARCHAR(20) NOT NULL,
    data_type VARCHAR(20) NOT NULL, -- 'kline', 'ticker', 'depth', 'trade'
    interval VARCHAR(5), -- Only for klines
    last_sync_time TIMESTAMP WITH TIME ZONE NOT NULL,
    last_data_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    error_message TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (symbol, data_type, COALESCE(interval, ''))
);

CREATE INDEX idx_sync_status_symbol ON sync_status(symbol, data_type);

-- Add retention policies (optional, adjust based on requirements)
-- Keep detailed 1m klines for 30 days, then downsample
-- SELECT add_retention_policy('klines', INTERVAL '90 days');

-- Keep ticker data for 90 days
-- SELECT add_retention_policy('tickers', INTERVAL '90 days');

-- Keep depth snapshots for 7 days (they're large)
-- SELECT add_retention_policy('depth_snapshots', INTERVAL '7 days');

-- Keep trades for 90 days
-- SELECT add_retention_policy('trades', INTERVAL '90 days');

-- Insert some default symbols (major pairs)
INSERT INTO symbols (symbol, base_asset, quote_asset, status, is_active) VALUES
    ('BTCUSDT', 'BTC', 'USDT', 'TRADING', true),
    ('ETHUSDT', 'ETH', 'USDT', 'TRADING', true),
    ('BNBUSDT', 'BNB', 'USDT', 'TRADING', true),
    ('ADAUSDT', 'ADA', 'USDT', 'TRADING', true),
    ('DOGEUSDT', 'DOGE', 'USDT', 'TRADING', true),
    ('XRPUSDT', 'XRP', 'USDT', 'TRADING', true),
    ('SOLUSDT', 'SOL', 'USDT', 'TRADING', true),
    ('DOTUSDT', 'DOT', 'USDT', 'TRADING', true),
    ('MATICUSDT', 'MATIC', 'USDT', 'TRADING', true),
    ('LINKUSDT', 'LINK', 'USDT', 'TRADING', true)
ON CONFLICT (symbol) DO NOTHING;

