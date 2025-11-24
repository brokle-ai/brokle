-- Add bloom filter index on service_name for faster service-filtered queries
ALTER TABLE traces ADD INDEX idx_service_name service_name TYPE bloom_filter(0.01) GRANULARITY 1
