BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_resource_interactions CASCADE;

CREATE MATERIALIZED VIEW mv_resource_interactions AS
WITH expanded_resources AS (
  SELECT
    f.id AS endpoint_id,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
         ELSE f.capability_fhir_version
    END AS fhir_version,

    -- Extract resource type from the JSONB structure
    resource_elem->>'type' AS resource_type,

    -- Extract individual operation names (this expands into multiple rows)
    COALESCE(interaction_elem->>'code', 'not specified') AS operation_name

  FROM fhir_endpoints_info f
  LEFT JOIN vendors v ON f.vendor_id = v.id

  -- Expand the "resource" array
  LEFT JOIN LATERAL json_array_elements((f.capability_statement->'rest')->0->'resource') resource_elem
    ON TRUE

	-- Expand the "interaction" array within each resource
  LEFT JOIN LATERAL json_array_elements(resource_elem->'interaction') interaction_elem
    ON TRUE
	
  WHERE f.requested_fhir_version = 'None'
),
aggregated_operations AS (
  SELECT
    vendor_name,
    fhir_version,
    resource_type,
	COUNT(DISTINCT endpoint_id) AS endpoint_count,
    -- Aggregate operations into an array
    ARRAY_AGG(DISTINCT operation_name) AS operations

  FROM expanded_resources
  GROUP BY vendor_name, fhir_version, resource_type
)
SELECT *
FROM aggregated_operations;

DROP INDEX IF EXISTS mv_resource_interactions_uniq;

CREATE UNIQUE INDEX mv_resource_interactions_uniq
  ON mv_resource_interactions (
    vendor_name,
    fhir_version,
    resource_type,
    endpoint_count,
    operations
  );

DROP INDEX IF EXISTS mv_resource_interactions_vendor_name_idx;

CREATE INDEX mv_resource_interactions_vendor_name_idx
  ON mv_resource_interactions (vendor_name);

DROP INDEX IF EXISTS mv_resource_interactions_fhir_version_idx;

CREATE INDEX mv_resource_interactions_fhir_version_idx
  ON mv_resource_interactions (fhir_version);

DROP INDEX IF EXISTS mv_resource_interactions_resource_type_idx;

CREATE INDEX mv_resource_interactions_resource_type_idx
  ON mv_resource_interactions (resource_type);

DROP INDEX IF EXISTS mv_resource_interactions_operations_idx;

CREATE INDEX mv_resource_interactions_operations_idx
  ON mv_resource_interactions USING GIN (operations);

COMMIT;