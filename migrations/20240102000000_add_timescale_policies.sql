-- +goose Up
-- This migration requires TimescaleDB extension to be installed
-- Run: CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE; before this migration

-- Convert telegrams table to hypertable (time-series optimization)
-- This enables automatic partitioning by time intervals
SELECT create_hypertable('telegrams', 'time', if_not_exists => TRUE);

-- Add 180-day retention policy
-- Automatically drops chunks older than 180 days
SELECT add_retention_policy('telegrams', INTERVAL '180 days', if_not_exists => TRUE);

-- Add compression for chunks older than 7 days
-- Reduces storage cost for historical data
ALTER TABLE telegrams SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'type,priority',
  timescaledb.compress_orderby = 'time DESC'
);

-- Schedule automatic compression for chunks older than 7 days
SELECT add_compression_policy('telegrams', INTERVAL '7 days', if_not_exists => TRUE);

-- +goose Down
-- Remove compression policy
SELECT remove_compression_policy('telegrams', if_exists => TRUE);

-- Remove retention policy
SELECT remove_retention_policy('telegrams', if_exists => TRUE);

-- Note: We don't revert the hypertable conversion as it may cause data loss
-- To fully revert, you would need to:
-- 1. Export data
-- 2. Drop the hypertable
-- 3. Recreate as regular table
-- 4. Re-import data

