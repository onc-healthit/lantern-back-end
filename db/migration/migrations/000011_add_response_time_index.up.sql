BEGIN;

CREATE INDEX metadata_response_time_idx ON fhir_endpoints_metadata(response_time_seconds);

COMMIT;