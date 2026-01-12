BEGIN;

-- Create table to store developers sharing the same service base URLs (list_sources)
-- This data is populated from CHPL's "Service Base URL List" CSV download
CREATE TABLE IF NOT EXISTS shared_list_sources (
    list_source VARCHAR(500) NOT NULL,
    developer_name VARCHAR(500) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (list_source, developer_name)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_shared_list_sources_list_source
ON shared_list_sources(list_source);

CREATE INDEX IF NOT EXISTS idx_shared_list_sources_updated_at
ON shared_list_sources(updated_at);

COMMIT;
