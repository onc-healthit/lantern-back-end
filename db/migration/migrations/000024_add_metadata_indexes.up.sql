BEGIN;

CREATE INDEX IF NOT EXISTS metadata_requested_version_idx ON fhir_endpoints_metadata(requested_fhir_version);
CREATE INDEX IF NOT EXISTS metadata_url_idx ON fhir_endpoints_metadata(url);

COMMIT;