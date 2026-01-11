-- Migration: add_contracts_expiring_index
-- Created: 2026-01-10T07:39:07+05:30

-- Optimized partial index for contract expiration worker query
-- Covers: WHERE status='active' AND end_date IS NOT NULL AND end_date <= ? ORDER BY end_date ASC
-- Uses partial index to only index active contracts with end dates, reducing index size
CREATE INDEX idx_contracts_expiring
    ON contracts(end_date ASC)
    WHERE status = 'active' AND end_date IS NOT NULL;
