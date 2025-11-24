-- Add service_name materialized column for OTLP compliance
-- service.name is the MOST CRITICAL resource attribute per OTLP spec
-- Materialized from metadata.resourceAttributes.service.name
-- Note: Bloom filter index added in separate migration due to ClickHouse multi-statement restrictions

ALTER TABLE traces ADD COLUMN
    service_name LowCardinality(String) MATERIALIZED
        JSONExtractString(metadata, 'resourceAttributes.service.name') CODEC(ZSTD(1))
