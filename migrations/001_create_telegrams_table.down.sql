-- Drop trigger and function
DROP TRIGGER IF EXISTS update_telegrams_updated_at ON telegrams;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_telegrams_time_type;
DROP INDEX IF EXISTS idx_telegrams_content_gin;
DROP INDEX IF EXISTS idx_telegrams_priority;
DROP INDEX IF EXISTS idx_telegrams_destination;
DROP INDEX IF EXISTS idx_telegrams_source;
DROP INDEX IF EXISTS idx_telegrams_flight_number;
DROP INDEX IF EXISTS idx_telegrams_type;
DROP INDEX IF EXISTS idx_telegrams_time;

-- Drop table
DROP TABLE IF EXISTS telegrams;

