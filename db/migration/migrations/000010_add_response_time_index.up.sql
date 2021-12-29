BEGIN;

CREATE INDEX IF NOT EXISTS metadata_response_time_idx ON fhir_endpoints_metadata(response_time_seconds);

COMMIT;