BEGIN;

-- Composite index covering the most common query pattern:
--   WHERE url = $1 AND requested_fhir_version = $2
-- Previously only separate single-column indexes existed, forcing the query planner
-- to pick one and filter by the other.
CREATE INDEX idx_fhir_endpoints_info_url_reqver
    ON fhir_endpoints_info (url, requested_fhir_version);

-- Same composite for the availability table, used by the
-- update_fhir_endpoint_availability_info trigger on every metadata insert/update.
CREATE INDEX idx_fhir_endpoints_availability_url_reqver
    ON fhir_endpoints_availability (url, requested_fhir_version);

COMMIT;
