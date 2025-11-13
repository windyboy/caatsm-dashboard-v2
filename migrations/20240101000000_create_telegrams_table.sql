-- +goose Up
-- Create telegrams table for storing aviation telegram messages
CREATE TABLE IF NOT EXISTS telegrams (
    message_id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    flight_number TEXT,
    source TEXT,
    destination TEXT,
    content TEXT,
    priority INT DEFAULT 3,
    raw_data TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_telegrams_time ON telegrams(time DESC);
CREATE INDEX IF NOT EXISTS idx_telegrams_type ON telegrams(type);
CREATE INDEX IF NOT EXISTS idx_telegrams_flight_number ON telegrams(flight_number);
CREATE INDEX IF NOT EXISTS idx_telegrams_source ON telegrams(source);
CREATE INDEX IF NOT EXISTS idx_telegrams_destination ON telegrams(destination);
CREATE INDEX IF NOT EXISTS idx_telegrams_priority ON telegrams(priority);
CREATE INDEX IF NOT EXISTS idx_telegrams_time_type ON telegrams(time DESC, type);

-- Create GIN index for full-text search on content
CREATE INDEX IF NOT EXISTS idx_telegrams_content_gin ON telegrams USING GIN(to_tsvector('english', content));

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $func$ BEGIN NEW.updated_at = NOW(); RETURN NEW; END; $func$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_telegrams_updated_at BEFORE UPDATE ON telegrams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
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

