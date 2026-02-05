BEGIN;

DROP INDEX IF EXISTS idx_mv_endpoint_resource_types_unique;

CREATE UNIQUE INDEX idx_mv_endpoint_resource_types_unique ON mv_endpoint_resource_types(endpoint_id, vendor_id, fhir_version, type);

COMMIT;