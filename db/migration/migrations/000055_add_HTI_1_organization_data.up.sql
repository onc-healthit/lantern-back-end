BEGIN;

DROP TABLE IF EXISTS fhir_endpoint_organization_active;

CREATE TABLE fhir_endpoint_organization_active (
	org_id INT,
	active VARCHAR(500)
);

DROP INDEX IF EXISTS idx_fhir_endpoint_organization_active_org_id;

CREATE INDEX idx_fhir_endpoint_organization_active_org_id ON fhir_endpoint_organization_active (org_id);

DROP TABLE IF EXISTS fhir_endpoint_organization_addresses;

CREATE TABLE fhir_endpoint_organization_addresses (
	org_id INT,
	address VARCHAR(500)
);

DROP INDEX IF EXISTS idx_fhir_endpoint_organization_addresses_org_id;

CREATE INDEX idx_fhir_endpoint_organization_addresses_org_id ON fhir_endpoint_organization_addresses (org_id);

DROP TABLE IF EXISTS fhir_endpoint_organization_identifiers;

CREATE TABLE fhir_endpoint_organization_identifiers (
	org_id INT,
	identifier VARCHAR(500)
);

DROP INDEX IF EXISTS idx_fhir_endpoint_organization_identifiers_org_id;

CREATE INDEX idx_fhir_endpoint_organization_identifiers_org_id ON fhir_endpoint_organization_identifiers (org_id);

DROP VIEW IF EXISTS joined_export_tables CASCADE;

CREATE or REPLACE VIEW joined_export_tables AS
SELECT endpts.url, endpts.list_source, endpt_orgnames.organization_names AS endpoint_names,
    endpt_orgnames.organization_ids AS endpoint_ids,
    vendors.name as vendor_name,
    endpts_info.tls_version, endpts_info.mime_types, endpts_metadata.http_response,
    endpts_metadata.response_time_seconds, endpts_metadata.smart_http_response, endpts_metadata.errors,
    EXISTS (SELECT 1 FROM fhir_endpoints_info WHERE capability_statement::jsonb != 'null' AND endpts.url = fhir_endpoints_info.url) as CAP_STAT_EXISTS,
    endpts_info.capability_fhir_version AS FHIR_VERSION,
    endpts_info.capability_statement->>'publisher' AS PUBLISHER,
    endpts_info.capability_statement->'software'->'name' AS SOFTWARE_NAME,
    endpts_info.capability_statement->'software'->'version' AS SOFTWARE_VERSION,
    endpts_info.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE,
    endpts_info.capability_statement->'format' AS FORMAT,
    endpts_info.capability_statement->>'kind' AS KIND,
    endpts_info.updated_at AS INFO_UPDATED, endpts_info.created_at AS INFO_CREATED,
    endpts_info.requested_fhir_version, endpts_metadata.availability
FROM fhir_endpoints AS endpts
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN (SELECT fom.id as id, array_agg(fo.organization_name) as organization_names, array_agg(fo.id) as organization_ids 
FROM fhir_endpoints AS fe, fhir_endpoint_organizations_map AS fom, fhir_endpoint_organizations AS fo
WHERE fe.id = fom.id AND fom.org_database_id = fo.id
GROUP BY fom.id) as endpt_orgnames ON endpts.id = endpt_orgnames.id;

CREATE or REPLACE VIEW endpoint_export AS
SELECT export_tables.url, export_tables.list_source, export_tables.endpoint_names,
    export_tables.endpoint_ids,
	export_tables.vendor_name,
    export_tables.tls_version, export_tables.mime_types, export_tables.http_response,
    export_tables.response_time_seconds, export_tables.smart_http_response, export_tables.errors,
    export_tables.cap_stat_exists,
    export_tables.fhir_version,
    export_tables.publisher,
    export_tables.software_name,
    export_tables.software_version,
    export_tables.software_releasedate,
    export_tables.format,
    export_tables.kind,
    export_tables.info_updated,
    export_tables.info_created,
    export_tables.requested_fhir_version,
    export_tables.availability,
    list_source_info.is_chpl
FROM joined_export_tables AS export_tables
LEFT JOIN list_source_info ON export_tables.list_source = list_source_info.list_source;

CREATE or REPLACE VIEW organization_location AS
    SELECT export_tables.url, export_tables.endpoint_names, export_tables.fhir_version,
    export_tables.requested_fhir_version, export_tables.vendor_name, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, orgs.npi_id as NPI_ID,
    links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
RIGHT JOIN joined_export_tables AS export_tables ON links.url = export_tables.url
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id
WHERE links.confidence > .97 AND orgs.Location->>'zipcode' IS NOT null;

CREATE MATERIALIZED VIEW mv_vendor_fhir_counts AS
WITH vendor_totals AS (
    -- Calculate total endpoints for each vendor
    SELECT 
        COALESCE(v.name, 'Unknown') AS vendor_name,
        COUNT(DISTINCT e.url) AS total_endpoints
    FROM endpoint_export e
    LEFT JOIN vendors v ON e.vendor_name = v.name
    GROUP BY COALESCE(v.name, 'Unknown')
),
vendor_rank AS (
    -- Rank vendors by total endpoints and determine top 10
    SELECT 
        vendor_name,
        total_endpoints,
        RANK() OVER (ORDER BY total_endpoints DESC) AS rank
    FROM vendor_totals
),
endpoint_counts_base AS (
    -- Get counts by vendor and FHIR version
    SELECT 
        COALESCE(v.name, 'Unknown') AS vendor_name,
        CASE
            WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
            -- Apply the dash rule: if there's a dash, trim after it
            WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
            -- If it's not in the valid list, mark as Unknown
            WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
            ELSE e.fhir_version
        END AS fhir_version,
        COUNT(DISTINCT e.url) AS n,
        -- Flag whether this vendor should be part of "Others"
        CASE 
            WHEN r.rank <= 10 THEN false
            WHEN t.total_endpoints < 50 THEN true
            ELSE false
        END AS is_other
    FROM endpoint_export e
    LEFT JOIN vendors v ON e.vendor_name = v.name
    JOIN vendor_totals t ON COALESCE(v.name, 'Unknown') = t.vendor_name
    JOIN vendor_rank r ON t.vendor_name = r.vendor_name
    GROUP BY 
        COALESCE(v.name, 'Unknown'), 
        CASE
            WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
            WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
            WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
            ELSE e.fhir_version
        END,
        r.rank,
        t.total_endpoints
),
others_aggregated AS (
    -- Aggregate "others" into one row per FHIR version
    SELECT 
        'Others' AS vendor_name,
        fhir_version,
        SUM(n) AS n,
        true AS is_other
    FROM endpoint_counts_base
    WHERE is_other = true
    GROUP BY fhir_version
),
non_others AS (
    -- Keep non-others as is
    SELECT
        vendor_name,
        fhir_version,
        n,
        is_other
    FROM endpoint_counts_base
    WHERE is_other = false
),
combined AS (
    -- Combine both sets
    SELECT * FROM others_aggregated
    UNION ALL
    SELECT * FROM non_others
)
SELECT 
    c.vendor_name,
    c.fhir_version,
    c.n,
    -- Add a sort order field
    CASE
        WHEN r.rank IS NOT NULL AND r.rank <= 10 THEN r.rank
        WHEN c.vendor_name = 'Others' THEN 9999  -- Others go at the bottom
        ELSE 1000 + COALESCE(r.rank, 0)  -- Keep remaining larger vendors in order
    END AS sort_order
FROM combined c
LEFT JOIN vendor_rank r ON c.vendor_name = r.vendor_name
ORDER BY sort_order, vendor_name, fhir_version;

CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);
CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);
CREATE INDEX idx_mv_vendor_fhir_counts_sort ON mv_vendor_fhir_counts(sort_order);

CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version, sort_order);

CREATE MATERIALIZED VIEW endpoint_export_mv AS
WITH endpoint_organizations AS (
    SELECT DISTINCT url, UNNEST(endpoint_names) AS endpoint_name
    FROM endpoint_export
),
grouped_organizations AS (
    SELECT url, 
           STRING_AGG(endpoint_name, '; ') AS endpoint_names 
    FROM endpoint_organizations
    WHERE endpoint_name IS NOT NULL AND endpoint_name <> 'NULL'
    GROUP BY url
),
processed_versions AS (
    SELECT 
        e.*,
        -- Step 1: Replace empty fhir_version with "No Cap Stat"
        CASE 
            WHEN e.fhir_version = '' THEN 'No Cap Stat'
            ELSE e.fhir_version
        END AS capability_fhir_version,
        -- Step 2: Extract version before "-" if present
        CASE 
            WHEN e.fhir_version = '' THEN 'No Cap Stat'
            WHEN POSITION('-' IN e.fhir_version) > 0 THEN SPLIT_PART(e.fhir_version, '-', 1)
            ELSE e.fhir_version
        END AS fhir_version_raw
    FROM endpoint_export e
)
SELECT 
    p.url, 
    p.list_source, 
    COALESCE(NULLIF(p.vendor_name, ''), 'Unknown') AS vendor_name,
    p.capability_fhir_version,
    -- Step 3: Use the fixed list of valid FHIR versions 
    CASE 
        WHEN p.capability_fhir_version = 'No Cap Stat' THEN 'No Cap Stat'  -- Ensure "No Cap Stat" is preserved
        WHEN p.fhir_version_raw IN ('No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', 
                                  '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', 
                                  '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', 
                                  '4.0.0', '4.0.1')
            THEN p.fhir_version_raw
        ELSE 'Unknown'  
    END AS fhir_version,
    p.tls_version,
    p.mime_types,
    p.http_response,
    p.response_time_seconds,
    p.smart_http_response,
    p.errors,
    p.cap_stat_exists,
    p.publisher,
    p.software_name,
    p.software_version,
    p.software_releasedate,
    REGEXP_REPLACE(p.format::TEXT, '[\[\]"]', '', 'g') AS format, 
    p.kind,
    p.info_updated,
    p.info_created,
    p.requested_fhir_version,
    p.availability,
    lsi.is_chpl,
    COALESCE(g.endpoint_names, '') AS endpoint_names
FROM processed_versions p
LEFT JOIN list_source_info lsi 
    ON p.list_source = lsi.list_source
LEFT JOIN grouped_organizations g 
    ON p.url = g.url;

-- Unique Index for refeshing the MV concurrently 
CREATE UNIQUE INDEX endpoint_export_mv_unique_idx ON endpoint_export_mv (url, list_source, vendor_name, fhir_version, info_updated);

--fhir_endpoint_comb_mv
CREATE MATERIALIZED VIEW fhir_endpoint_comb_mv AS 
SELECT 
    ROW_NUMBER() OVER () AS id,
    t.url,
    t.endpoint_names,
    t.info_created,
    t.info_updated,
    t.list_source,
    t.vendor_name,
    t.capability_fhir_version,
    t.fhir_version,
    t.format,
    t.http_response,
    t.response_time_seconds,
    t.smart_http_response,
    t.errors,
    t.availability,
    t.kind,
    t.requested_fhir_version,
    t.is_chpl,
    t.status,
    t.cap_stat_exists
FROM (
    SELECT DISTINCT ON (e.url, e.vendor_name, e.fhir_version, e.http_response, e.requested_fhir_version)
        e.url,
        e.endpoint_names,
        e.info_created,
        e.info_updated,
        e.list_source,
        e.vendor_name,
        e.capability_fhir_version,
        e.fhir_version,
        e.format,
        e.http_response,
        e.response_time_seconds,
        e.smart_http_response,
        e.errors,
        e.availability,
        e.kind,
        e.requested_fhir_version,
        lsi.is_chpl,
        CASE 
            WHEN e.http_response = 200 THEN CONCAT('Success: ', e.http_response, ' - ', r.code_label)
            WHEN e.http_response IS NULL OR e.http_response = 0 THEN 'Failure: 0 - NA'
            ELSE CONCAT('Failure: ', e.http_response, ' - ', r.code_label)
        END AS status,
        LOWER(CASE 
            WHEN e.kind != 'instance' THEN 'true*'::TEXT  
            ELSE e.cap_stat_exists::TEXT
        END) AS cap_stat_exists
    FROM endpoint_export_mv e
    LEFT JOIN mv_http_responses r ON e.http_response = r.http_code
    LEFT JOIN list_source_info lsi ON e.list_source = lsi.list_source
    ORDER BY e.url, e.vendor_name, e.fhir_version, e.http_response, e.requested_fhir_version
) t;

--Unique index for refreshing the MV concurrently
CREATE UNIQUE INDEX fhir_endpoint_comb_mv_unique_idx ON fhir_endpoint_comb_mv (id, url, list_source);

--selected_fhir_endpoints_mv
CREATE MATERIALIZED VIEW selected_fhir_endpoints_mv AS
SELECT 
    ROW_NUMBER() OVER () AS id,  -- Generate a unique sequential ID
    e.url,
    e.endpoint_names,
    e.info_created,
    e.info_updated,
    e.list_source,
    e.vendor_name,
    e.capability_fhir_version,
    e.fhir_version,
    e.format,
    e.http_response,
    e.response_time_seconds,
    e.smart_http_response,
    e.errors,
    e.availability * 100 AS availability,
    e.kind,
    e.requested_fhir_version,
    lsi.is_chpl,
    e.status,
    e.cap_stat_exists,
    
    -- Generate URL modal link
    CONCAT('<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop-up modal containing additional information for this endpoint." 
            onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" 
            onclick="Shiny.setInputValue(''endpoint_popup'',''', e.url, '&&', e.requested_fhir_version, ''',{priority: ''event''});">', e.url, '</a>') 
    AS "urlModal",

    -- Generate Condensed Endpoint Names
    CASE 
        WHEN e.endpoint_names IS NOT NULL 
             AND array_length(string_to_array(e.endpoint_names, ';'), 1) > 5
        THEN CONCAT(
            array_to_string(ARRAY(SELECT unnest(string_to_array(e.endpoint_names, ';')) LIMIT 5), '; '),
            '; <a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop-up modal containing the endpoint''s entire list of API information source names." 
                onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" 
                onclick="Shiny.setInputValue(''show_details'',''', e.url, ''',{priority: ''event''});"> Click For More... </a>'
        )
        ELSE e.endpoint_names
    END AS condensed_endpoint_names

FROM fhir_endpoint_comb_mv e
LEFT JOIN list_source_info lsi 
    ON e.list_source = lsi.list_source;

-- Create a unique composite index including the new id column
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_mv_unique ON selected_fhir_endpoints_mv(id, url, requested_fhir_version);

-- Create single column indexes to improve filtering performance
CREATE INDEX idx_selected_fhir_endpoints_mv_fhir_version ON selected_fhir_endpoints_mv(fhir_version);
CREATE INDEX idx_selected_fhir_endpoints_mv_vendor_name ON selected_fhir_endpoints_mv(vendor_name);
CREATE INDEX idx_selected_fhir_endpoints_mv_availability ON selected_fhir_endpoints_mv(availability);
CREATE INDEX idx_selected_fhir_endpoints_mv_is_chpl ON selected_fhir_endpoints_mv(is_chpl);

CREATE MATERIALIZED VIEW mv_contacts_info AS
WITH contact_info_extracted AS (
  SELECT DISTINCT
    url,
    json_array_elements((capability_statement->>'contact')::json)->>'name' as contact_name,
    json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'system' as contact_type,
    json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'value' as contact_value,
    CAST(NULLIF(json_array_elements((json_array_elements((capability_statement->>'contact')::json)->>'telecom')::json)->>'rank', '') AS INTEGER) as contact_preference
  FROM fhir_endpoints_info
  WHERE capability_statement::jsonb != 'null' AND requested_fhir_version = 'None'
),
endpoint_details AS (
  SELECT DISTINCT -- Added DISTINCT to eliminate potential duplication
    url,
    -- Fix for handling Unknown vendor - make sure empty or NULL is replaced with 'Unknown'
    CASE 
      WHEN vendor_name IS NULL OR vendor_name = '' THEN 'Unknown' 
      ELSE vendor_name 
    END AS vendor_name,
    CASE 
      WHEN fhir_version = '' OR fhir_version IS NULL THEN 'No Cap Stat'
      WHEN position('-' in fhir_version) > 0 THEN substring(fhir_version from 1 for position('-' in fhir_version) - 1)
      WHEN fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
      ELSE fhir_version
    END AS fhir_version,
    requested_fhir_version
  FROM endpoint_export
  WHERE requested_fhir_version = 'None'
),
endpoint_names_grouped AS (
  SELECT 
    url, 
    string_agg(DISTINCT endpoint_names_list, ';') AS endpoint_names -- Added DISTINCT to avoid duplications
  FROM (
    SELECT DISTINCT url, UNNEST(endpoint_names) as endpoint_names_list 
    FROM endpoint_export 
    WHERE requested_fhir_version = 'None'
    ORDER BY endpoint_names_list
  ) AS unnested
  GROUP BY url
),
-- First, get URLs with contact info
urls_with_contacts AS (
  SELECT DISTINCT url
  FROM contact_info_extracted
),
-- Then, get URLs without contact info
urls_without_contacts AS (
  SELECT DISTINCT e.url
  FROM endpoint_details e
  LEFT JOIN urls_with_contacts c ON e.url = c.url
  WHERE c.url IS NULL
),
-- Combine contact data
joined_with_contacts AS (
  SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    eng.endpoint_names,
    e.requested_fhir_version,
    c.contact_name,
    c.contact_type,
    c.contact_value,
    COALESCE(c.contact_preference, 999) AS contact_preference,
    TRUE AS has_contact,
    MD5(CONCAT(
      e.url, 
      COALESCE(c.contact_name, ''), 
      COALESCE(c.contact_type, ''), 
      COALESCE(c.contact_value, ''),
      COALESCE(c.contact_preference::text, '999'),
      COALESCE(random()::text, '')  -- Add randomness to handle duplicates
    )) AS unique_hash
  FROM 
    endpoint_details e
  INNER JOIN 
    urls_with_contacts uc ON e.url = uc.url
  LEFT JOIN 
    endpoint_names_grouped eng ON e.url = eng.url
  LEFT JOIN 
    contact_info_extracted c ON e.url = c.url
),
-- Handle URLs without contacts
joined_without_contacts AS (
  SELECT 
    e.url,
    e.vendor_name,
    e.fhir_version,
    eng.endpoint_names,
    e.requested_fhir_version,
    NULL AS contact_name,
    NULL AS contact_type,
    NULL AS contact_value,
    999 AS contact_preference,
    FALSE AS has_contact,
    MD5(CONCAT(
      e.url, 
      'no_contact',
      COALESCE(random()::text, '')  -- Add randomness to handle duplicates
    )) AS unique_hash
  FROM 
    endpoint_details e
  INNER JOIN 
    urls_without_contacts nc ON e.url = nc.url
  LEFT JOIN 
    endpoint_names_grouped eng ON e.url = eng.url
)
-- Combine both sets
SELECT * FROM joined_with_contacts
UNION ALL
SELECT * FROM joined_without_contacts
ORDER BY 
  url, 
  contact_preference;

CREATE UNIQUE INDEX idx_mv_contacts_info_unique ON mv_contacts_info(unique_hash);
CREATE INDEX idx_mv_contacts_info_url ON mv_contacts_info(url);
CREATE INDEX idx_mv_contacts_info_fhir_version ON mv_contacts_info(fhir_version);
CREATE INDEX idx_mv_contacts_info_vendor_name ON mv_contacts_info(vendor_name);
CREATE INDEX idx_mv_contacts_info_has_contact ON mv_contacts_info(has_contact);
CREATE INDEX idx_mv_contacts_info_contact_preference ON mv_contacts_info(contact_preference);

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_endpoint_list_organizations
AS
SELECT DISTINCT
    endpoint_export.url,
    COALESCE(
        NULLIF(
            btrim(
                regexp_replace(name_id.cleaned_name, '\s+', ' ', 'g')
            ), 
        ''), 
    'Unknown') AS organization_name,
    
    COALESCE(
        name_id.cleaned_id::text,  -- Just cast to text
        'Unknown'
    ) AS organization_id,
    
    CASE
        WHEN endpoint_export.fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
        ELSE endpoint_export.fhir_version
    END AS fhir_version,
    
    COALESCE(endpoint_export.vendor_name, 'Unknown'::character varying) AS vendor_name

FROM
    endpoint_export
LEFT JOIN LATERAL (
    SELECT
        name_elem AS cleaned_name,
        id_elem AS cleaned_id
    FROM
        unnest(endpoint_export.endpoint_names, endpoint_export.endpoint_ids) AS u(name_elem, id_elem)
) AS name_id ON TRUE

ORDER BY
    organization_name
WITH DATA;

 -- Create indexes for endpoint list organizations materialized view
CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, organization_id);
CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);
CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);
CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);

CREATE MATERIALIZED VIEW security_endpoints_mv AS
SELECT 
    ROW_NUMBER() OVER () AS id,
    e.url,
    REPLACE(
        REPLACE(
            REPLACE(
                REPLACE(e.endpoint_names::TEXT, '{', ''), 
                '}', ''
            ), 
            '","', '; '
        ),
        '"', ''
    ) AS organization_names,
    COALESCE(e.vendor_name, 'Unknown') AS vendor_name,
    CASE 
        WHEN e.fhir_version = '' THEN 'No Cap Stat'
        ELSE e.fhir_version 
    END AS capability_fhir_version,
    e.tls_version,
    codes.code,
    CASE 
        -- First transform empty to "No Cap Stat"
        WHEN e.fhir_version = '' THEN 'No Cap Stat'
        -- Then handle version with dash
        WHEN e.fhir_version LIKE '%-%' THEN 
            CASE 
                WHEN SPLIT_PART(e.fhir_version, '-', 1) IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
                THEN SPLIT_PART(e.fhir_version, '-', 1)
                ELSE 'Unknown'
            END
        -- Handle regular versions
        WHEN e.fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
        THEN e.fhir_version
        ELSE 'Unknown'
    END AS fhir_version_final
FROM endpoint_export e
JOIN fhir_endpoints_info f ON e.url = f.url
JOIN LATERAL (
    SELECT json_array_elements(json_array_elements(f.capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code
) codes ON true
WHERE f.requested_fhir_version = 'None';

--indexing 
CREATE INDEX idx_security_endpoints_url ON security_endpoints_mv (url);
CREATE INDEX idx_security_endpoints_fhir_version ON security_endpoints_mv (fhir_version_final);
CREATE INDEX idx_security_endpoints_vendor_name ON security_endpoints_mv (vendor_name);
CREATE INDEX idx_security_endpoints_code ON security_endpoints_mv (code);
--unique index
CREATE UNIQUE INDEX idx_unique_security_endpoints ON security_endpoints_mv (id, url, vendor_name, code);

CREATE MATERIALIZED VIEW selected_security_endpoints_mv AS
SELECT 
    se.id,
    se.url,
    se.organization_names,
    se.vendor_name,
    se.capability_fhir_version,
    se.fhir_version_final AS fhir_version,
    se.tls_version,
    se.code,
    -- Create the condensed_organization_names with the modal link for endpoints with more than 5 organizations
    CASE 
        WHEN se.organization_names IS NOT NULL AND 
             array_length(string_to_array(se.organization_names, ';'), 1) > 5 
        THEN 
            CONCAT(
                array_to_string(
                    ARRAY(
                        SELECT unnest(string_to_array(se.organization_names, ';')) 
                        LIMIT 5
                    ), 
                    '; '
                ),
                '; <a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing the endpoint''s entire list of API information source names." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''show_details'',''', 
                se.url, 
                ''',{priority: ''event''});"> Click For More... </a>'
            )
        ELSE 
            se.organization_names 
    END AS condensed_organization_names,
    
    -- Create the URL with modal functionality
    CONCAT(
        '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',''', 
        se.url, 
        '&&None'',{priority: ''event''});">', 
        se.url, 
        '</a>'
    ) AS url_modal
FROM 
    security_endpoints_mv se;

-- Add indexing for better performance
CREATE INDEX idx_selected_security_endpoints_fhir_version ON selected_security_endpoints_mv (fhir_version);
CREATE INDEX idx_selected_security_endpoints_vendor_name ON selected_security_endpoints_mv (vendor_name);
CREATE INDEX idx_selected_security_endpoints_code ON selected_security_endpoints_mv (code);
-- Create a unique composite index
CREATE UNIQUE INDEX idx_unique_selected_security_endpoints ON selected_security_endpoints_mv (id, url, code);

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