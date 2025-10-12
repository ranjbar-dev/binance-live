-- Initialize sync status for all active symbols and data types
-- This creates initial sync status records for kline intervals

-- Initialize sync status for kline intervals
INSERT INTO sync_status (symbol, data_type, interval, last_sync_time, last_data_time, status)
SELECT 
    s.symbol,
    'kline' as data_type,
    intervals.interval,
    0 as last_sync_time,
    0 as last_data_time,
    'pending' as status
FROM symbols s
CROSS JOIN (
    VALUES 
        ('1m'),
        ('15m'), 
        ('1h'),
        ('4h'),
        ('1d')
) AS intervals(interval)
WHERE s.is_active = true
ON CONFLICT (symbol, data_type, interval) DO NOTHING;

-- Initialize sync status for other data types (without intervals)
INSERT INTO sync_status (symbol, data_type, interval, last_sync_time, last_data_time, status)
SELECT 
    s.symbol,
    data_types.data_type,
    '' as interval,
    0 as last_sync_time,
    0 as last_data_time,
    'pending' as status
FROM symbols s
CROSS JOIN (
    VALUES 
        ('ticker'),
        ('depth'),
        ('trade')
) AS data_types(data_type)
WHERE s.is_active = true
ON CONFLICT (symbol, data_type, interval) DO NOTHING;
