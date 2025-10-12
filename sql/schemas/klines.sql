-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Table to store kline/candlestick data
CREATE TABLE IF NOT EXISTS klines (
    symbol VARCHAR(20) NOT NULL,
    interval VARCHAR(5) NOT NULL,
    open_time BIGINT NOT NULL,
    close_time BIGINT NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    high_price DECIMAL(20, 8) NOT NULL,
    low_price DECIMAL(20, 8) NOT NULL,
    close_price DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    quote_volume DECIMAL(20, 8) NOT NULL,
    trades_count INTEGER NOT NULL,
    taker_buy_volume DECIMAL(20, 8) NOT NULL,
    taker_buy_quote_volume DECIMAL(20, 8) NOT NULL,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()) * 1000,
    PRIMARY KEY (symbol, interval, open_time)
);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('klines', 'open_time', if_not_exists => TRUE);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval ON klines(symbol, interval, open_time DESC);
