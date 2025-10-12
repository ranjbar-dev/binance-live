-- Sample depth snapshot data (this would typically be populated by the application)
-- This file is intentionally left mostly empty as depth data is real-time data
-- that should be populated by the live data collection system

-- You can add sample depth snapshot data here if needed for testing:
-- INSERT INTO depth_snapshots (
--     symbol, timestamp, last_update_id, bids, asks
-- ) VALUES (
--     'BTCUSDT', EXTRACT(EPOCH FROM NOW()) * 1000, 12345678,
--     '[["35190.00", "0.5"], ["35180.00", "1.2"], ["35170.00", "0.8"]]'::jsonb,
--     '[["35210.00", "0.3"], ["35220.00", "0.7"], ["35230.00", "1.1"]]'::jsonb
-- );

-- Note: Bids and asks are stored as JSONB arrays of [price, quantity] pairs
-- Note: Use EXTRACT(EPOCH FROM NOW()) * 1000 for current timestamp in milliseconds
