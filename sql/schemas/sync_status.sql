-- Table to track sync status for each symbol
CREATE TABLE IF NOT EXISTS sync_status (
    symbol VARCHAR(20) NOT NULL,
    data_type VARCHAR(20) NOT NULL, -- 'kline', 'ticker', 'depth', 'trade'
    interval VARCHAR(5) NOT NULL DEFAULT '', -- Only for klines, empty string for others
    last_sync_time BIGINT NOT NULL,
    last_data_time BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    error_message TEXT,
    updated_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()) * 1000,
    PRIMARY KEY (symbol, data_type, interval)
);

CREATE INDEX IF NOT EXISTS idx_sync_status_symbol ON sync_status(symbol, data_type);
