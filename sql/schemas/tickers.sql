-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table to store ticker/price data
CREATE TABLE IF NOT EXISTS tickers (
    symbol VARCHAR(20) NOT NULL,
    timestamp BIGINT NOT NULL,
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
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()) * 1000,
    PRIMARY KEY (symbol, timestamp)
);

-- Convert to hypertable
SELECT create_hypertable('tickers', 'timestamp', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_tickers_symbol ON tickers(symbol, timestamp DESC);
