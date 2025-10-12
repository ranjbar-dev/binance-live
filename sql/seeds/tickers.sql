-- Sample ticker data (this would typically be populated by the application)
-- This file is intentionally left mostly empty as ticker data is real-time data
-- that should be populated by the live data collection system

-- You can add sample ticker data here if needed for testing:
-- INSERT INTO tickers (
--     symbol, timestamp, price, bid_price, bid_qty, ask_price, ask_qty,
--     volume_24h, quote_volume_24h, price_change_24h, price_change_percent_24h,
--     high_24h, low_24h, trades_count_24h
-- ) VALUES (
--     'BTCUSDT', EXTRACT(EPOCH FROM NOW()) * 1000, 35200.00, 35190.00, 0.5,
--     35210.00, 0.3, 50000.0, 1760000000.0, 1200.00, 3.53,
--     36000.00, 34000.00, 125000
-- );

-- Note: Use EXTRACT(EPOCH FROM NOW()) * 1000 for current timestamp in milliseconds
