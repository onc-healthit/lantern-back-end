BEGIN;

-- Create materialized view for endpoint_organization_tbl
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_organization_tbl CASCADE;
CREATE MATERIALIZED VIEW mv_endpoint_organization_tbl AS
 SELECT sub.url,
    array_agg(sub.endpoint_names_list ORDER BY sub.endpoint_names_list) AS endpoint_names_list
   FROM ( SELECT DISTINCT endpoint_export.url,
            unnest(endpoint_export.endpoint_names) AS endpoint_names_list
           FROM endpoint_export) sub
  GROUP BY sub.url
  ORDER BY sub.url;

-- Create indexes for mv_endpoint_organization_tbl
CREATE UNIQUE INDEX idx_mv_endpoint_list_org_url_uniq ON mv_endpoint_organization_tbl(url);


-- Create materialized view for endpoint_export_tbl
DROP MATERIALIZED VIEW IF EXISTS mv_endpoint_export_tbl CASCADE;
CREATE MATERIALIZED VIEW mv_endpoint_export_tbl AS
WITH base AS (
  SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    e.format,
    e.endpoint_names,
    e.http_response,
    e.smart_http_response,
    o.endpoint_names_list
  FROM endpoint_export e
  LEFT JOIN mv_endpoint_organization_tbl o 
    ON e.url::text = o.url::text
),
mutated AS (
  SELECT
    url,
    -- Replace empty vendor_name with NULL then later to "Unknown"
    CASE WHEN vendor_name = '' THEN NULL ELSE vendor_name END AS vendor_name,
    CASE WHEN fhir_version = '' THEN 'No Cap Stat' ELSE fhir_version END AS capability_fhir_version,
    -- Compute fhir_version: if contains a hyphen, take the part before it; else, keep as-is.
    CASE 
      WHEN fhir_version = '' THEN 'No Cap Stat'
      WHEN fhir_version LIKE '%-%' THEN split_part(fhir_version, '-', 1)
      ELSE fhir_version 
    END AS fhir_version,
    format,
    endpoint_names,
    http_response,
    smart_http_response,
    endpoint_names_list
  FROM base
),
validated AS (
  SELECT
    url,
    COALESCE(vendor_name, 'Unknown') AS vendor_name,
    capability_fhir_version,
    CASE 
      WHEN fhir_version IN (
           'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2',
           '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0',
           '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0',
           '3.5.0', '3.5a.0', '4.0.0', '4.0.1'
         )
      THEN fhir_version
      ELSE 'Unknown'
    END AS fhir_version,
    format,
    endpoint_names,
    http_response,
    smart_http_response,
    endpoint_names_list
  FROM mutated
),
cleaned AS (
  SELECT
  	row_number() OVER () AS mv_id,
    url,
    vendor_name,
    capability_fhir_version,
    fhir_version,
    -- Clean the format column by removing quotes and square brackets.
    regexp_replace(format::text, '("|\\[|\\])', '', 'g') AS format,
    -- Fully clean endpoint_names_list:
    regexp_replace(
      regexp_replace(
        regexp_replace(
          regexp_replace(CAST(endpoint_names_list AS text), '^c\\(|\\)$', '', 'g'),
          '(", )', '"; ', 'g'
        ),
        'NULL', '', 'g'
      ),
      '"', '', 'g'
    ) AS endpoint_names,
    http_response,
    smart_http_response
  FROM validated
)
SELECT *
FROM cleaned
ORDER BY url;

-- Create indexes for mv_endpoint_export_tbl
CREATE UNIQUE INDEX idx_mv_endpoint_export_tbl_unique_id ON mv_endpoint_export_tbl(mv_id);
CREATE INDEX idx_mv_endpoint_export_tbl_vendor ON mv_endpoint_export_tbl (vendor_name);
CREATE INDEX idx_mv_endpoint_export_tbl_fhir ON mv_endpoint_export_tbl (fhir_version);
CREATE INDEX idx_mv_endpoint_export_tbl_vendor_fhir ON mv_endpoint_export_tbl (vendor_name, fhir_version);

-- Create materialized view for http_pct
DROP MATERIALIZED VIEW IF EXISTS mv_http_pct CASCADE;
CREATE MATERIALIZED VIEW mv_http_pct AS
WITH grouped AS (
  SELECT
    f.id,
    f.url,
    e.http_response,
    e.vendor_name,
    e.fhir_version,
    CAST(e.http_response AS text) AS code,
    COUNT(*) AS cnt
  FROM fhir_endpoints_info f
  LEFT JOIN mv_endpoint_export_tbl e ON f.url = e.url
  WHERE f.requested_fhir_version = 'None'
  GROUP BY f.id, f.url, e.http_response, e.vendor_name, e.fhir_version
)
SELECT
  row_number() OVER () AS mv_id,
  id,
  url,
  http_response,
  COALESCE(vendor_name, 'Unknown') AS vendor_name,
  fhir_version,
  code,
  cnt * 100.0 / SUM(cnt) OVER (PARTITION BY id) AS Percentage
FROM grouped;

-- Create indexes for mv_http_pct
CREATE UNIQUE INDEX idx_mv_http_pct_unique_id ON mv_http_pct(mv_id);
CREATE INDEX idx_mv_http_pct_http_response ON mv_http_pct (http_response);
CREATE INDEX idx_mv_http_pct_vendor ON mv_http_pct (vendor_name);
CREATE INDEX idx_mv_http_pct_fhir ON mv_http_pct (fhir_version);
CREATE INDEX idx_mv_http_pct_vendor_fhir ON mv_http_pct (vendor_name, fhir_version);

-- Create materialized view for well_known_endpoints
DROP MATERIALIZED VIEW IF EXISTS mv_well_known_endpoints CASCADE;
CREATE MATERIALIZED VIEW mv_well_known_endpoints AS

WITH base AS (
         SELECT e.url,
            array_to_string(e.endpoint_names, ';'::text) AS organization_names,
            COALESCE(e.vendor_name, 'Unknown'::character varying) AS vendor_name,
                CASE
                    WHEN e.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
                    ELSE e.fhir_version
                END AS capability_fhir_version
           FROM endpoint_export e
             LEFT JOIN fhir_endpoints_info f ON e.url::text = f.url::text
             LEFT JOIN fhir_endpoints_metadata m ON f.metadata_id = m.id
             LEFT JOIN vendors v ON f.vendor_id = v.id
          WHERE m.smart_http_response = 200 AND f.requested_fhir_version::text = 'None'::text AND jsonb_typeof(f.smart_response::jsonb) = 'object'::text
        )
 SELECT 
   	row_number() OVER () AS mv_id,
	base.url,
    regexp_replace(regexp_replace(regexp_replace(base.organization_names, '[{}]'::text, ''::text, 'g'::text), '","'::text, '; '::text, 'g'::text), '"'::text, ''::text, 'g'::text) AS organization_names,
    base.vendor_name,
    base.capability_fhir_version,
        CASE
            WHEN
            CASE
                WHEN base.capability_fhir_version::text ~~ '%-%'::text THEN split_part(base.capability_fhir_version::text, '-'::text, 1)::character varying
                ELSE base.capability_fhir_version
            END::text = ANY (ARRAY['No Cap Stat'::character varying, '0.4.0'::character varying, '0.5.0'::character varying, '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0.1'::character varying]::text[]) THEN
            CASE
                WHEN base.capability_fhir_version::text ~~ '%-%'::text THEN split_part(base.capability_fhir_version::text, '-'::text, 1)::character varying
                ELSE base.capability_fhir_version
            END
            ELSE 'Unknown'::character varying
        END AS fhir_version
   FROM base;

-- Create indexes for mv_well_known_endpoints
CREATE UNIQUE INDEX idx_mv_well_known_unique_id ON mv_well_known_endpoints(mv_id);
CREATE INDEX idx_mv_well_known_vendor ON mv_well_known_endpoints(vendor_name);
CREATE INDEX idx_mv_well_known_fhir ON mv_well_known_endpoints(fhir_version);
CREATE INDEX idx_mv_well_known_vendor_fhir ON mv_well_known_endpoints(vendor_name, fhir_version);

-- Create materialized view for well_known_no_doc
DROP MATERIALIZED VIEW IF EXISTS mv_well_known_no_doc CASCADE;
CREATE MATERIALIZED VIEW mv_well_known_no_doc AS

WITH base AS (
	 SELECT f.id,
		e.url,
		f.vendor_id,
		e.endpoint_names AS organization_names,
		e.vendor_name,
		e.fhir_version,
		m.smart_http_response,
		f.smart_response
	   FROM endpoint_export e
		 LEFT JOIN fhir_endpoints_info f ON e.url::text = f.url::text
		 LEFT JOIN fhir_endpoints_metadata m ON f.metadata_id = m.id
		 LEFT JOIN vendors v ON f.vendor_id = v.id
	  WHERE m.smart_http_response = 200 AND f.requested_fhir_version::text = 'None'::text AND jsonb_typeof(f.smart_response::jsonb) <> 'object'::text
	)
SELECT 
	row_number() OVER () AS mv_id,
	base.id,
	base.url,
	base.vendor_id,
	base.organization_names,
	COALESCE(base.vendor_name, 'Unknown'::character varying) AS vendor_name,
	base.smart_http_response,
	base.smart_response,
	CASE
		WHEN
		CASE
			WHEN base.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
			WHEN base.fhir_version::text ~~ '%-%'::text THEN split_part(base.fhir_version::text, '-'::text, 1)::character varying
			ELSE base.fhir_version
		END::text = ANY (ARRAY['No Cap Stat'::character varying, '0.4.0'::character varying, '0.5.0'::character varying, '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0.1'::character varying]::text[]) THEN
		CASE
			WHEN base.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
			WHEN base.fhir_version::text ~~ '%-%'::text THEN split_part(base.fhir_version::text, '-'::text, 1)::character varying
			ELSE base.fhir_version
		END
		ELSE 'Unknown'::character varying
	END AS fhir_version
FROM base;

-- Create indexes for mv_well_known_no_doc
CREATE UNIQUE INDEX idx_mv_well_known_no_doc_unique_id ON mv_well_known_no_doc(mv_id);
CREATE INDEX idx_mv_well_known_no_doc_url ON mv_well_known_no_doc(url);
CREATE INDEX idx_mv_well_known_no_doc_vendor ON mv_well_known_no_doc(vendor_name);
CREATE INDEX idx_mv_well_known_no_doc_fhir ON mv_well_known_no_doc(fhir_version);
CREATE INDEX idx_mv_well_known_no_doc_vendor_fhir ON mv_well_known_no_doc(vendor_name, fhir_version);

-- Create materialized view for smart_response_capabilities
DROP MATERIALIZED VIEW IF EXISTS mv_smart_response_capabilities CASCADE;
CREATE MATERIALIZED VIEW mv_smart_response_capabilities AS

WITH original AS (
 SELECT 
 	f.id,
    m.smart_http_response,
    COALESCE(v.name, 'Unknown'::character varying) AS vendor_name,
        CASE
            WHEN f.capability_fhir_version::text = ''::text THEN 'No Cap Stat'::text
            WHEN f.capability_fhir_version::text ~~ '%-%'::text THEN
            CASE
                WHEN split_part(f.capability_fhir_version::text, '-'::text, 1) = ANY (ARRAY['No Cap Stat'::text, '0.4.0'::text, '0.5.0'::text, '1.0.0'::text, '1.0.1'::text, '1.0.2'::text, '1.1.0'::text, '1.2.0'::text, '1.4.0'::text, '1.6.0'::text, '1.8.0'::text, '3.0.0'::text, '3.0.1'::text, '3.0.2'::text, '3.2.0'::text, '3.3.0'::text, '3.5.0'::text, '3.5a.0'::text, '4.0.0'::text, '4.0.1'::text]) THEN split_part(f.capability_fhir_version::text, '-'::text, 1)
                ELSE 'Unknown'::text
            END
            WHEN f.capability_fhir_version::text = ANY (ARRAY['No Cap Stat'::character varying::text, '0.4.0'::character varying::text, '0.5.0'::character varying::text, '1.0.0'::character varying::text, '1.0.1'::character varying::text, '1.0.2'::character varying::text, '1.1.0'::character varying::text, '1.2.0'::character varying::text, '1.4.0'::character varying::text, '1.6.0'::character varying::text, '1.8.0'::character varying::text, '3.0.0'::character varying::text, '3.0.1'::character varying::text, '3.0.2'::character varying::text, '3.2.0'::character varying::text, '3.3.0'::character varying::text, '3.5.0'::character varying::text, '3.5a.0'::character varying::text, '4.0.0'::character varying::text, '4.0.1'::character varying::text]) THEN f.capability_fhir_version::text
            ELSE 'Unknown'::text
        END AS fhir_version,
    json_array_elements_text(f.smart_response -> 'capabilities'::text) AS capability
   FROM fhir_endpoints_info f
     JOIN vendors v ON f.vendor_id = v.id
     JOIN fhir_endpoints_metadata m ON f.metadata_id = m.id
  WHERE f.requested_fhir_version::text = 'None'::text AND m.smart_http_response = 200)
SELECT row_number() OVER () AS mv_id,
       original.*
FROM original;

-- Create indexes for mv_smart_response_capabilities
CREATE UNIQUE INDEX idx_mv_smart_response_capabilities_unique_id ON mv_smart_response_capabilities(mv_id);
CREATE INDEX idx_mv_smart_response_capabilities_id ON mv_smart_response_capabilities (id);
CREATE INDEX idx_mv_smart_response_capabilities_vendor ON mv_smart_response_capabilities (vendor_name);
CREATE INDEX idx_mv_smart_response_capabilities_fhir ON mv_smart_response_capabilities (fhir_version);
CREATE INDEX idx_mv_smart_response_capabilities_capability ON mv_smart_response_capabilities (capability);
CREATE INDEX idx_mv_smart_response_capabilities_vendor_fhir ON mv_smart_response_capabilities (vendor_name, fhir_version);
CREATE INDEX idx_mv_smart_response_capabilities_capability_fhir ON mv_smart_response_capabilities (capability, fhir_version);

-- Create materialized view for selected_endpoints
DROP MATERIALIZED VIEW IF EXISTS mv_selected_endpoints CASCADE;
CREATE MATERIALIZED VIEW mv_selected_endpoints AS
WITH original AS (
 SELECT
 	DISTINCT mv_well_known_endpoints.url,
        CASE
            WHEN mv_well_known_endpoints.organization_names IS NULL OR mv_well_known_endpoints.organization_names = ''::text THEN mv_well_known_endpoints.organization_names
            ELSE
            CASE
                WHEN cardinality(string_to_array(mv_well_known_endpoints.organization_names, ';'::text)) > 5 THEN (((array_to_string(( SELECT array_agg(t.elem) AS array_agg
                   FROM unnest(string_to_array(mv_well_known_endpoints.organization_names, ';'::text)) WITH ORDINALITY t(elem, ord)
                  WHERE t.ord <= 5), ';'::text) || '; '::text) || '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing the endpoint''s entire list of API information source names." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click();}})(event)" onclick="Shiny.setInputValue(''show_details'','''::text) || mv_well_known_endpoints.url::text) || ''',{priority: ''event''});"> Click For More... </a>'::text
                ELSE mv_well_known_endpoints.organization_names
            END
        END AS condensed_organization_names,
    mv_well_known_endpoints.vendor_name,
    mv_well_known_endpoints.capability_fhir_version
 FROM mv_well_known_endpoints)
 SELECT 
   row_number() OVER (ORDER BY url) AS mv_id,
   *
 FROM original;

-- Create indexes for mv_selected_endpoints
CREATE UNIQUE INDEX idx_mv_selected_endpoints_unique_id ON mv_selected_endpoints(mv_id);
CREATE INDEX idx_mv_selected_endpoints_vendor ON mv_selected_endpoints(vendor_name);
CREATE INDEX idx_mv_selected_endpoints_fhir ON mv_selected_endpoints(capability_fhir_version);
CREATE INDEX idx_mv_selected_endpoints_vendor_fhir ON mv_selected_endpoints(vendor_name, capability_fhir_version);

COMMIT;