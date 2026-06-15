BEGIN;

DROP MATERIALIZED VIEW IF EXISTS mv_well_known_endpoints CASCADE;

-- Create materialized view for well_known_endpoints
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
            END::text = ANY (ARRAY[
                'No Cap Stat'::character varying, '0.4.0'::character varying, '0.5.0'::character varying, 
                '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, 
                '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, 
                '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, 
                '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, 
                '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, 
                '4.0.0'::character varying, '4.0.1'::character varying, '4.1.0'::character varying, 
                '4.3.0'::character varying, '4.2.0'::character varying, '4.4.0'::character varying, 
                '4.5.0'::character varying, '4.6.0'::character varying, '5.0.0'::character varying
            ]::text[]) THEN
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
		END::text = ANY (ARRAY[
            'No Cap Stat'::character varying, '0.4.0'::character varying, '0.5.0'::character varying, 
            '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, 
            '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, 
            '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, 
            '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, 
            '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, 
            '4.0.0'::character varying, '4.0.1'::character varying, '4.1.0'::character varying, 
            '4.3.0'::character varying, '4.2.0'::character varying, '4.4.0'::character varying, 
            '4.5.0'::character varying, '4.6.0'::character varying, '5.0.0'::character varying
        ]::text[]) THEN
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

COMMIT;