-- Add bloom filter index on service_name for 5-10x faster service-filtered queries
ALTER TABLE spans ADD INDEX idx_service_name service_name TYPE bloom_filter(0.01) GRANULARITY 1
