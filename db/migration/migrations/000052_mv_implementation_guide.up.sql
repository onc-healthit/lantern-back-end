BEGIN;

-- Create materialized view for implementation_guide
DROP MATERIALIZED VIEW IF EXISTS mv_implementation_guide CASCADE;
CREATE MATERIALIZED VIEW mv_implementation_guide AS 

SELECT
  f.url AS url,
  CASE 
    WHEN split_part(
           CASE 
             WHEN f.capability_fhir_version = '' THEN 'No Cap Stat' 
             ELSE f.capability_fhir_version 
           END, '-', 1)
         IN ('No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1')
      THEN split_part(
             CASE 
               WHEN f.capability_fhir_version = '' THEN 'No Cap Stat' 
               ELSE f.capability_fhir_version 
             END, '-', 1)
      ELSE 'Unknown'
  END AS fhir_version,
  json_array_elements_text(f.capability_statement::json#>'{implementationGuide}') AS implementation_guide,
  COALESCE(vendors.name, 'Unknown') AS vendor_name
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.requested_fhir_version = 'None';

-- Create indexes for mv_implementation_guide
CREATE UNIQUE INDEX idx_mv_implementation_guide_unique ON mv_implementation_guide(url, fhir_version, implementation_guide, vendor_name);
CREATE INDEX idx_mv_implementation_guide_vendor ON mv_implementation_guide(vendor_name);
CREATE INDEX idx_mv_implementation_guide_fhir ON mv_implementation_guide(fhir_version);

COMMIT;