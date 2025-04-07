CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $fhir_endpoints_info_history$
    BEGIN
        --
        -- Create a row in fhir_endpoints_info_history to reflect the operation performed on fhir_endpoints_info,
        -- make use of the special variable TG_OP to work out the operation.
        --
        IF (TG_OP = 'DELETE') THEN
            INSERT INTO fhir_endpoints_info_history SELECT 'D', now(), user, OLD.*;
            RETURN OLD;
        ELSIF (TG_OP = 'UPDATE') THEN
            INSERT INTO fhir_endpoints_info_history SELECT 'U', now(), user, NEW.*;
            RETURN NEW;
        ELSIF (TG_OP = 'INSERT') THEN
            INSERT INTO fhir_endpoints_info_history SELECT 'I', now(), user, NEW.*;
            RETURN NEW;
        END IF;
        RETURN NULL; -- result is ignored since this is an AFTER trigger
    END;
$fhir_endpoints_info_history$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_fhir_endpoint_availability_info() RETURNS TRIGGER AS $fhir_endpoints_availability$
    DECLARE
        okay_count       bigint;
        all_count        bigint;
    BEGIN
        --
        -- Create or update a row in fhir_endpoint_availabilty with new total http and 200 http count 
        -- when an endpoint is inserted or updated in fhir_endpoint_info. Also calculate new 
        -- endpoint availability precentage
        SELECT http_200_count, http_all_count INTO okay_count, all_count FROM fhir_endpoints_availability WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
        IF  NOT FOUND THEN
            IF NEW.http_response = 200 THEN
                INSERT INTO fhir_endpoints_availability(url, http_200_count, http_all_count, requested_fhir_version) VALUES (NEW.url, 1, 1, NEW.requested_fhir_version);
                NEW.availability = 1.00;
                RETURN NEW;
            ELSE
                INSERT INTO fhir_endpoints_availability(url, http_200_count, http_all_count, requested_fhir_version) VALUES (NEW.url, 0, 1, NEW.requested_fhir_version);
                NEW.availability = 0.00;
                RETURN NEW;
            END IF;
        ELSE
            IF NEW.http_response = 200 THEN
                UPDATE fhir_endpoints_availability SET http_200_count = okay_count + 1.0, http_all_count = all_count + 1.0 WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
                NEW.availability := (okay_count + 1.0) / (all_count + 1.0);
                RETURN NEW;
            ELSE
                UPDATE fhir_endpoints_availability SET http_all_count = all_count + 1.0 WHERE url = NEW.url AND requested_fhir_version = NEW.requested_fhir_version;
                NEW.availability := (okay_count) / (all_count + 1.0);
                RETURN NEW;
            END IF;
        END IF;
    END;
$fhir_endpoints_availability$ LANGUAGE plpgsql;

CREATE TABLE npi_organizations (
    id                          SERIAL PRIMARY KEY,
    npi_id                      VARCHAR(500) UNIQUE,
    name                        VARCHAR(500),
    secondary_name              VARCHAR(500),
    location                    JSONB,
    taxonomy                    VARCHAR(500), -- Taxonomy code mapping: http://www.wpc-edi.com/reference/codelists/healthcare/health-care-provider-taxonomy-code-set/
    normalized_name             VARCHAR(500),
    normalized_secondary_name   VARCHAR(500),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE npi_contacts (
    id                                  SERIAL PRIMARY KEY,
    npi_id                              VARCHAR(500),
    endpoint_type                       VARCHAR(500),
    endpoint_type_description           VARCHAR(500),
    endpoint                            VARCHAR(500),
    valid_url                           BOOLEAN,
    affiliation                         VARCHAR(500),
    endpoint_description                VARCHAR(500),
    affiliation_legal_business_name     VARCHAR(500),
    use_code                            VARCHAR(500),
    use_description                     VARCHAR(500),
    other_use_description               VARCHAR(500),
    content_type                        VARCHAR(500),
    content_description                 VARCHAR(500),
    other_content_description           VARCHAR(500),
    location                            JSONB,
    created_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vendors (
    id                      SERIAL PRIMARY KEY,
    name                    VARCHAR(500) UNIQUE,
    developer_code          VARCHAR(500) UNIQUE,
    url                     VARCHAR(500),
    location                JSONB,
    status                  VARCHAR(500),
    last_modified_in_chpl   TIMESTAMPTZ,
    chpl_id                 INTEGER UNIQUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE healthit_products (
    id                      SERIAL PRIMARY KEY,
    name                    VARCHAR(500),
    version                 VARCHAR(500),
    vendor_id               INT REFERENCES vendors(id) ON DELETE CASCADE,
    location                JSONB,
    authorization_standard  VARCHAR(500),
    api_syntax              VARCHAR(500),
    api_url                 VARCHAR(500),
    certification_criteria  JSONB,
    certification_status    VARCHAR(500),
    certification_date      DATE,
    certification_edition   VARCHAR(500),
    last_modified_in_chpl   DATE,
    chpl_id                 VARCHAR(500),
    practice_type           VARCHAR(500),
    acb                     VARCHAR(500),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT healthit_product_info UNIQUE(name, version)
);

CREATE TABLE healthit_products_map (
    id SERIAL,
    healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL,
    CONSTRAINT unique_id_healthit_product UNIQUE(id, healthit_product_id)
);

CREATE TABLE certification_criteria (
    id                        SERIAL PRIMARY KEY,
    certification_id          INTEGER,
	cerification_number       VARCHAR(500),
	title                     VARCHAR(500),
	certification_edition_id  INTEGER,
	certification_edition     VARCHAR(500),
	description               VARCHAR(500),
	removed                   BOOLEAN,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fhir_endpoints (
    id                      SERIAL PRIMARY KEY,
    url                     VARCHAR(500),
    list_source             VARCHAR(500),
    versions_response       JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fhir_endpoints_unique UNIQUE(url, list_source)
);

CREATE TABLE fhir_endpoint_organizations (
    id                      SERIAL PRIMARY KEY,
    organization_name       VARCHAR(500),
    organization_zipcode    VARCHAR(500),
    organization_npi_id    VARCHAR(500),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fhir_endpoint_organizations_map (
    id INT REFERENCES fhir_endpoints(id) ON DELETE SET NULL,
    org_database_id INT REFERENCES fhir_endpoint_organizations(id) ON DELETE SET NULL
);

CREATE TABLE fhir_endpoints_metadata (
    id                      SERIAL PRIMARY KEY,
    url                     VARCHAR(500),
    http_response           INTEGER,
    availability            DECIMAL(5,4),
    errors                  VARCHAR(500),
    response_time_seconds   DECIMAL(7,4),
    smart_http_response     INTEGER,
    requested_fhir_version VARCHAR(500) DEFAULT 'None',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE validation_results (
    id                      SERIAL PRIMARY KEY
);

CREATE TABLE fhir_endpoints_info (
    id                      SERIAL PRIMARY KEY,
    healthit_mapping_id     INT, -- should link to healthit_products_map(id). not using 'reference' because the referenced id might have multiple entries and thus is not a primary key
    vendor_id               INT REFERENCES vendors(id) ON DELETE SET NULL, 
    url                     VARCHAR(500),
    tls_version             VARCHAR(500),
    mime_types              VARCHAR(500)[],
    capability_statement    JSON,
    validation_result_id    INT REFERENCES validation_results(id) ON DELETE SET NULL,
    included_fields         JSONB,
    operation_resource      JSONB,
    supported_profiles      JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    smart_response          JSON,
    metadata_id             INT REFERENCES fhir_endpoints_metadata(id) ON DELETE SET NULL,
    requested_fhir_version  VARCHAR(500),
    capability_fhir_version VARCHAR(500),
    CONSTRAINT fhir_endpoints_info_unique UNIQUE(url, requested_fhir_version)
);

CREATE TABLE fhir_endpoints_info_history (
    operation               CHAR(1) NOT NULL,
    entered_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id                 VARCHAR(500),
    id                      INT, -- should link to fhir_endpoints_info(id). not using 'reference' because if the original is deleted, we still want the historical copies to remain and keep the ID so they can be linked to one another.
    healthit_mapping_id     INT, -- should link to healthit_products_map(id). not using 'reference' because if the referenced product is deleted, we still want the historical copies to retain the ID.
    vendor_id               INT,  -- should link to vendor_id(id). not using 'reference' because if the referenced vendor is deleted, we still want the historical copies to retain the ID.
    url                     VARCHAR(500),
    tls_version             VARCHAR(500),
    mime_types              VARCHAR(500)[],
    capability_statement    JSON,
    validation_result_id    INT REFERENCES validation_results(id) ON DELETE SET NULL,
    included_fields         JSONB,
    operation_resource      JSONB,
    supported_profiles      JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    smart_response          JSON, 
    metadata_id             INT REFERENCES fhir_endpoints_metadata(id) ON DELETE SET NULL,
    requested_fhir_version  VARCHAR(500),
    capability_fhir_version VARCHAR(500)
);

CREATE TABLE endpoint_organization (
    url                     VARCHAR(500),
    organization_npi_id     VARCHAR(500),
    confidence              NUMERIC (5, 3),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT endpoint_org PRIMARY KEY (url, organization_npi_id)
);

CREATE TABLE list_source_info (
    list_source            VARCHAR(500),
    is_chpl                BOOLEAN 
);

CREATE TABLE product_criteria (
    healthit_product_id      INT REFERENCES healthit_products(id) ON DELETE CASCADE,
    certification_id         INTEGER,
    certification_number     VARCHAR(500),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_crit  PRIMARY KEY (healthit_product_id, certification_id)
);

CREATE TABLE fhir_endpoints_availability (
    url             VARCHAR(500),
    http_200_count       BIGINT,
    http_all_count       BIGINT,
    requested_fhir_version  VARCHAR(500) DEFAULT 'None',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE validations (
    rule_name               VARCHAR(500),
    valid                   BOOLEAN,
    expected                VARCHAR(500),
    actual                  VARCHAR(500),
    comment                 VARCHAR(500),
    reference               VARCHAR(500),
    implementation_guide    VARCHAR(500),
    validation_result_id    INT REFERENCES validation_results(id) ON DELETE SET NULL
);

CREATE TABLE info_history_pruning_metadata (
    id                                  SERIAL PRIMARY KEY,
    started_on                          timestamp with time zone NOT NULL DEFAULT now(),
    ended_on                            timestamp with time zone,
    successful                          boolean NOT NULL DEFAULT false,
    num_rows_processed                  integer NOT NULL DEFAULT 0,
    num_rows_pruned                     integer NOT NULL DEFAULT 0,
    query_int_start_date                timestamp with time zone NOT NULL,
    query_int_end_date                  timestamp with time zone NOT NULL
);

CREATE TRIGGER set_timestamp_fhir_endpoints
BEFORE UPDATE ON fhir_endpoints
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_fhir_endpoint_organizations
BEFORE UPDATE ON fhir_endpoint_organizations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_npi_organization
BEFORE UPDATE ON npi_organizations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_vendors
BEFORE UPDATE ON vendors
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_healthit_products
BEFORE UPDATE ON healthit_products
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_certification_criteria
BEFORE UPDATE ON certification_criteria
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_fhir_endpoints_info
BEFORE UPDATE ON fhir_endpoints_info
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_fhir_endpoints_metadata
BEFORE UPDATE ON fhir_endpoints_metadata
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_endpoint_organization
BEFORE UPDATE ON endpoint_organization
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_product_criteria
BEFORE UPDATE ON product_criteria
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_fhir_endpoint_availability
BEFORE UPDATE ON fhir_endpoints_availability
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- captures history for the fhir_endpoint_info table
CREATE TRIGGER add_fhir_endpoint_info_history_trigger
AFTER INSERT OR UPDATE OR DELETE on fhir_endpoints_info
FOR EACH ROW
WHEN (current_setting('metadata.setting', 't') IS NULL OR current_setting('metadata.setting', 't') = 'FALSE')
EXECUTE PROCEDURE add_fhir_endpoint_info_history();

-- increments total number of times http status returned for endpoint 
CREATE TRIGGER update_fhir_endpoint_availability_trigger
BEFORE INSERT OR UPDATE on fhir_endpoints_metadata
FOR EACH ROW
EXECUTE PROCEDURE update_fhir_endpoint_availability_info();

CREATE or REPLACE VIEW joined_export_tables AS
SELECT endpts.url, endpts.list_source, endpt_orgnames.organization_names AS endpoint_names,
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
LEFT JOIN (SELECT fom.id as id, array_agg(fo.organization_name) as organization_names 
FROM fhir_endpoints AS fe, fhir_endpoint_organizations_map AS fom, fhir_endpoint_organizations AS fo
WHERE fe.id = fom.id AND fom.org_database_id = fo.id
GROUP BY fom.id) as endpt_orgnames ON endpts.id = endpt_orgnames.id;

CREATE or REPLACE VIEW endpoint_export AS
SELECT export_tables.url, export_tables.list_source, export_tables.endpoint_names,
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

CREATE INDEX fhir_endpoints_url_idx ON fhir_endpoints (url);
CREATE INDEX fhir_endpoints_info_url_idx ON fhir_endpoints_info (url);
CREATE INDEX fhir_endpoints_info_history_url_idx ON fhir_endpoints_info_history (url);
CREATE INDEX endpoint_organization_url_idx ON endpoint_organization (url);

CREATE INDEX fhir_endpoint_organizations_id ON fhir_endpoint_organizations (id);
CREATE INDEX fhir_endpoint_organizations_name ON fhir_endpoint_organizations (organization_name);
CREATE INDEX fhir_endpoint_organizations_zipcode ON fhir_endpoint_organizations (organization_zipcode);
CREATE INDEX fhir_endpoint_organizations_npi_id ON fhir_endpoint_organizations (organization_npi_id);

CREATE INDEX vendor_id_idx ON vendors (id);
CREATE INDEX fhir_endpoints_info_vendor_id_idx ON fhir_endpoints_info (vendor_id);
CREATE INDEX fhir_endpoints_info_history_vendor_id_idx ON fhir_endpoints_info_history (vendor_id);

CREATE INDEX npi_organizations_npi_id_idx ON npi_organizations (npi_id);
CREATE INDEX endpoint_organization_npi_id_idx ON endpoint_organization (organization_npi_id);

CREATE INDEX vendor_name_idx ON vendors (name);
CREATE INDEX implementation_guide_idx ON fhir_endpoints_info ((capability_statement->>'implementationGuide'));
CREATE INDEX field_idx ON fhir_endpoints_info ((included_fields->> 'Field'));
CREATE INDEX exists_idx ON fhir_endpoints_info ((included_fields->> 'Exists'));
CREATE INDEX extension_idx ON fhir_endpoints_info ((included_fields->> 'Extension'));

CREATE INDEX resource_type_idx ON fhir_endpoints_info (((capability_statement::json#>'{rest,0,resource}') ->> 'type'));

CREATE INDEX capstat_url_idx ON fhir_endpoints_info ((capability_statement->>'url'));
CREATE INDEX capstat_version_idx ON fhir_endpoints_info ((capability_statement->>'version'));
CREATE INDEX capstat_name_idx ON fhir_endpoints_info ((capability_statement->>'name'));
CREATE INDEX capstat_title_idx ON fhir_endpoints_info ((capability_statement->>'title'));
CREATE INDEX capstat_date_idx ON fhir_endpoints_info ((capability_statement->>'date'));
CREATE INDEX capstat_publisher_idx ON fhir_endpoints_info ((capability_statement->>'publisher'));
CREATE INDEX capstat_description_idx ON fhir_endpoints_info ((capability_statement->>'description'));
CREATE INDEX capstat_purpose_idx ON fhir_endpoints_info ((capability_statement->>'purpose'));
CREATE INDEX capstat_copyright_idx ON fhir_endpoints_info ((capability_statement->>'copyright'));

CREATE INDEX capstat_software_name_idx ON fhir_endpoints_info ((capability_statement->'software'->>'name'));
CREATE INDEX capstat_software_version_idx ON fhir_endpoints_info ((capability_statement->'software'->>'version'));
CREATE INDEX capstat_software_releaseDate_idx ON fhir_endpoints_info ((capability_statement->'software'->>'releaseDate'));
CREATE INDEX capstat_implementation_description_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'description'));
CREATE INDEX capstat_implementation_url_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'url'));
CREATE INDEX capstat_implementation_custodian_idx ON fhir_endpoints_info ((capability_statement->'implementation'->>'custodian'));

CREATE INDEX capability_fhir_version_idx ON fhir_endpoints_info (capability_fhir_version);
CREATE INDEX requested_fhir_version_idx ON fhir_endpoints_info (requested_fhir_version);

CREATE INDEX security_code_idx ON fhir_endpoints_info ((capability_statement::json#>'{rest,0,security,service}'->'coding'->>'code'));
CREATE INDEX security_service_idx ON fhir_endpoints_info ((capability_statement::json#>'{rest,0,security}' -> 'service' ->> 'text'));

CREATE INDEX smart_capabilities_idx ON fhir_endpoints_info ((smart_response->>'capabilities'));

CREATE INDEX location_zipcode_idx ON npi_organizations ((location->>'zipcode'));

CREATE INDEX info_metadata_id_idx ON fhir_endpoints_info (metadata_id);
CREATE INDEX info_history_metadata_id_idx ON fhir_endpoints_info_history (metadata_id);
CREATE INDEX metadata_id_idx ON fhir_endpoints_metadata (id);

CREATE INDEX healthit_product_name_version_idx ON healthit_products (name, version);
CREATE INDEX metadata_response_time_idx ON fhir_endpoints_metadata(response_time_seconds);
CREATE INDEX metadata_requested_version_idx ON fhir_endpoints_metadata(requested_fhir_version);
CREATE INDEX metadata_url_idx ON fhir_endpoints_metadata(url);

-- LANTERN-759
CREATE INDEX validations_val_res_id_idx ON validations (validation_result_id);
CREATE INDEX fhir_endpoints_info_validation_result_id_idx ON fhir_endpoints_info (validation_result_id); 
CREATE INDEX fhir_endpoints_info_history_entered_at_idx ON fhir_endpoints_info_history (entered_at);
CREATE INDEX fhir_endpoints_info_history_operation_idx ON fhir_endpoints_info_history (operation);
CREATE INDEX fhir_endpoints_info_history_requested_fhir_version_idx ON fhir_endpoints_info_history (requested_fhir_version);
CREATE INDEX healthit_products_certification_status_idx ON healthit_products (certification_status);
CREATE INDEX healthit_products_chpl_id_idx ON healthit_products (chpl_id);
CREATE INDEX fhir_endpoint_organizations_map_id_idx ON fhir_endpoint_organizations_map (id);
CREATE INDEX fhir_endpoint_organizations_map_org_database_id_idx ON fhir_endpoint_organizations_map (org_database_id);

-- LANTERN-835

CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        (SELECT count(DISTINCT url) FROM fhir_endpoints) AS all_endpoints,
        (SELECT count(DISTINCT url) 
         FROM fhir_endpoints_info 
         WHERE requested_fhir_version = 'None') AS indexed_endpoints
)
SELECT 
    now() AS aggregation_date,
    totals.all_endpoints,
    totals.indexed_endpoints,
    greatest(totals.all_endpoints - totals.indexed_endpoints, 0) AS nonindexed_endpoints,
    (SELECT latest_metadata.last_updated FROM latest_metadata) AS last_updated
FROM totals;

CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

CREATE MATERIALIZED VIEW mv_response_tally AS
WITH response_counts AS (
    SELECT 
        fem.http_response,
        count(*) AS response_count
    FROM fhir_endpoints_info fei
    JOIN fhir_endpoints_metadata fem 
        ON fei.metadata_id = fem.id
    WHERE fei.requested_fhir_version = 'None'
    GROUP BY fem.http_response
)
SELECT 
    COALESCE(SUM(
        CASE 
            WHEN http_response = 200 THEN response_count 
            ELSE 0 
        END), 0) AS http_200,
    COALESCE(SUM(
        CASE 
            WHEN http_response <> 200 THEN response_count 
            ELSE 0 
        END), 0) AS http_non200,
    COALESCE(SUM(
        CASE 
            WHEN http_response = 404 THEN response_count 
            ELSE 0 
        END), 0) AS http_404,
    COALESCE(SUM(
        CASE 
            WHEN http_response = 503 THEN response_count 
            ELSE 0 
        END), 0) AS http_503
FROM response_counts;

CREATE UNIQUE INDEX idx_mv_response_tally_http_code ON mv_response_tally(http_200);

CREATE MATERIALIZED VIEW mv_vendor_fhir_counts AS
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
    CASE
        WHEN COALESCE(v.name, 'Unknown') = 'Allscripts' THEN 'Allscripts'
        WHEN COALESCE(v.name, 'Unknown') = 'CareEvolution, Inc.' THEN 'CareEvolution'
        WHEN COALESCE(v.name, 'Unknown') = 'Cerner Corporation' THEN 'Cerner'
        WHEN COALESCE(v.name, 'Unknown') = 'Epic Systems Corporation' THEN 'Epic'
        WHEN COALESCE(v.name, 'Unknown') = 'Medical Information Technology, Inc. (MEDITECH)' THEN 'MEDITECH'
        WHEN COALESCE(v.name, 'Unknown') = 'Microsoft Corporation' THEN 'Microsoft'
        WHEN COALESCE(v.name, 'Unknown') = 'Unknown' THEN 'Unknown'
        ELSE COALESCE(v.name, 'Unknown')
    END AS short_name
FROM endpoint_export e
LEFT JOIN vendors v ON e.vendor_name = v.name
GROUP BY 
    COALESCE(v.name, 'Unknown'), 
    CASE
        WHEN e.fhir_version IS NULL OR trim(e.fhir_version) = '' THEN 'No Cap Stat'
        WHEN position('-' in e.fhir_version) > 0 THEN substring(e.fhir_version, 1, position('-' in e.fhir_version) - 1)
        WHEN e.fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown'
        ELSE e.fhir_version
    END,
    CASE
        WHEN COALESCE(v.name, 'Unknown') = 'Allscripts' THEN 'Allscripts'
        WHEN COALESCE(v.name, 'Unknown') = 'CareEvolution, Inc.' THEN 'CareEvolution'
        WHEN COALESCE(v.name, 'Unknown') = 'Cerner Corporation' THEN 'Cerner'
        WHEN COALESCE(v.name, 'Unknown') = 'Epic Systems Corporation' THEN 'Epic'
        WHEN COALESCE(v.name, 'Unknown') = 'Medical Information Technology, Inc. (MEDITECH)' THEN 'MEDITECH'
        WHEN COALESCE(v.name, 'Unknown') = 'Microsoft Corporation' THEN 'Microsoft'
        WHEN COALESCE(v.name, 'Unknown') = 'Unknown' THEN 'Unknown'
        ELSE COALESCE(v.name, 'Unknown')
    END
ORDER BY 
    vendor_name, fhir_version;

-- Add indexes to improve query performance
CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);
CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);
CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version);

-- LANTERN-831
CREATE MATERIALIZED VIEW mv_http_responses AS
WITH response_by_vendor AS (
    SELECT
        CASE
            WHEN v.name IS NULL OR v.name = '' THEN 'Unknown'
            ELSE v.name
        END AS vendor_name,
        m.http_response AS http_code,
        CASE 
            WHEN m.http_response = 100 THEN 'Continue'
            WHEN m.http_response = 101 THEN 'Switching Protocols'
            WHEN m.http_response = 102 THEN 'Processing'
            WHEN m.http_response = 103 THEN 'Early Hints'
            WHEN m.http_response = 200 THEN 'OK'
            WHEN m.http_response = 201 THEN 'Created'
            WHEN m.http_response = 202 THEN 'Accepted'
            WHEN m.http_response = 203 THEN 'Non-Authoritative Information'
            WHEN m.http_response = 204 THEN 'No Content'
            WHEN m.http_response = 205 THEN 'Reset Content'
            WHEN m.http_response = 206 THEN 'Partial Content'
            WHEN m.http_response = 207 THEN 'Multi-Status'
            WHEN m.http_response = 208 THEN 'Already Reported'
            WHEN m.http_response = 226 THEN 'IM Used'
            WHEN m.http_response = 300 THEN 'Multiple Choices'
            WHEN m.http_response = 301 THEN 'Moved Permanently'
            WHEN m.http_response = 302 THEN 'Found'
            WHEN m.http_response = 303 THEN 'See Other'
            WHEN m.http_response = 304 THEN 'Not Modified'
            WHEN m.http_response = 305 THEN 'Use Proxy'
            WHEN m.http_response = 306 THEN 'Switch Proxy'
            WHEN m.http_response = 307 THEN 'Temporary Redirect'
            WHEN m.http_response = 308 THEN 'Permanent Redirect'
            WHEN m.http_response = 400 THEN 'Bad Request'
            WHEN m.http_response = 401 THEN 'Unauthorized'
            WHEN m.http_response = 402 THEN 'Payment Required'
            WHEN m.http_response = 403 THEN 'Forbidden'
            WHEN m.http_response = 404 THEN 'Not Found'
            WHEN m.http_response = 405 THEN 'Method Not Allowed'
            WHEN m.http_response = 406 THEN 'Not Acceptable'
            WHEN m.http_response = 407 THEN 'Proxy Authentication Required'
            WHEN m.http_response = 408 THEN 'Request Timeout'
            WHEN m.http_response = 409 THEN 'Conflict'
            WHEN m.http_response = 410 THEN 'Gone'
            WHEN m.http_response = 411 THEN 'Length Required'
            WHEN m.http_response = 412 THEN 'Precondition Failed'
            WHEN m.http_response = 413 THEN 'Payload Too Large'
            WHEN m.http_response = 414 THEN 'Request URI Too Long'
            WHEN m.http_response = 415 THEN 'Unsupported Media Type'
            WHEN m.http_response = 416 THEN 'Requested Range Not Satisfiable'
            WHEN m.http_response = 417 THEN 'Expectation Failed'
            WHEN m.http_response = 418 THEN 'I''m a teapot'
            WHEN m.http_response = 421 THEN 'Misdirected Request'
            WHEN m.http_response = 422 THEN 'Unprocessable Entity'
            WHEN m.http_response = 423 THEN 'Locked'
            WHEN m.http_response = 424 THEN 'Failed Dependency'
            WHEN m.http_response = 425 THEN 'Too Early'
            WHEN m.http_response = 426 THEN 'Upgrade Required'
            WHEN m.http_response = 428 THEN 'Precondition Required'
            WHEN m.http_response = 429 THEN 'Too Many Requests'
            WHEN m.http_response = 431 THEN 'Request Header Fields Too Large'
            WHEN m.http_response = 451 THEN 'Unavailable for Legal Reasons'
            WHEN m.http_response = 500 THEN 'Internal Server Error'
            WHEN m.http_response = 501 THEN 'Not Implemented'
            WHEN m.http_response = 502 THEN 'Bad Gateway'
            WHEN m.http_response = 503 THEN 'Service Unavailable'
            WHEN m.http_response = 504 THEN 'Gateway Timeout'
            WHEN m.http_response = 505 THEN 'HTTP Version Not Supported'
            WHEN m.http_response = 506 THEN 'Variant Also Negotiates'
            WHEN m.http_response = 507 THEN 'Insufficient Storage'
            WHEN m.http_response = 508 THEN 'Loop Detected'
            WHEN m.http_response = 509 THEN 'Bandwidth Limit Exceeded'
            WHEN m.http_response = 510 THEN 'Not Extended'
            WHEN m.http_response = 511 THEN 'Network Authentication Required'
            ELSE 'Other'
        END AS code_label,
        COUNT(DISTINCT f.url) AS count_endpoints
    FROM fhir_endpoints_info f
    LEFT JOIN vendors v
           ON f.vendor_id = v.id
    LEFT JOIN fhir_endpoints_metadata m
           ON f.metadata_id = m.id
    WHERE m.http_response IS NOT NULL
      AND f.requested_fhir_version = 'None'
    GROUP BY v.name, m.http_response
),
response_all_devs AS (
    SELECT
        'ALL_DEVELOPERS' AS vendor_name,
        http_code,
        code_label,
        SUM(count_endpoints) AS count_endpoints
    FROM response_by_vendor
    GROUP BY http_code, code_label
)
SELECT 
    now() AS aggregation_date,
    vendor_name,
    http_code,
    code_label,
    count_endpoints
FROM response_by_vendor

UNION ALL

SELECT
    now() AS aggregation_date,
    vendor_name,
    http_code,
    code_label,
    count_endpoints
FROM response_all_devs;

CREATE UNIQUE INDEX mv_http_responses_uniq
  ON mv_http_responses (aggregation_date, vendor_name, http_code);

CREATE INDEX mv_http_responses_vendor_name_idx
  ON mv_http_responses (vendor_name);

-- LANTERN-832
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

CREATE UNIQUE INDEX mv_resource_interactions_uniq
  ON mv_resource_interactions (
    vendor_name,
    fhir_version,
    resource_type,
    endpoint_count,
    operations
  );

CREATE INDEX mv_resource_interactions_vendor_name_idx
  ON mv_resource_interactions (vendor_name);

CREATE INDEX mv_resource_interactions_fhir_version_idx
  ON mv_resource_interactions (fhir_version);

CREATE INDEX mv_resource_interactions_resource_type_idx
  ON mv_resource_interactions (resource_type);

CREATE INDEX mv_resource_interactions_operations_idx
  ON mv_resource_interactions USING GIN (operations);


--LANTERN-848
CREATE MATERIALIZED VIEW get_capstat_values_mv AS
WITH valid_fhir_versions AS (
    -- Dynamically extract all distinct FHIR versions from the dataset
    SELECT DISTINCT 
        CASE 
            WHEN capability_fhir_version LIKE '%-%' THEN SPLIT_PART(capability_fhir_version, '-', 1)
            ELSE capability_fhir_version
        END AS version
    FROM fhir_endpoints_info
    WHERE capability_fhir_version IS NOT NULL
)
SELECT 
    f.id AS endpoint_id,
    f.vendor_id,
    COALESCE(vendors.name, 'Unknown') AS vendor_name,
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat' 
        ELSE f.capability_fhir_version 
    END AS fhir_version,
    -- Extract the major version dynamically
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
        ELSE f.capability_fhir_version 
    END AS raw_filter_fhir_version,
    -- Check dynamically against extracted valid FHIR versions
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN (
            CASE 
                WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
                ELSE f.capability_fhir_version 
            END
        ) IN (SELECT version FROM valid_fhir_versions) 
        THEN (
            CASE 
                WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
                ELSE f.capability_fhir_version 
            END
        ) 
        ELSE 'Unknown' 
    END AS filter_fhir_version,
    f.capability_statement->>'url' AS url,
    f.capability_statement->>'version' AS version,
    f.capability_statement->>'name' AS name,
    f.capability_statement->>'title' AS title,
    f.capability_statement->>'date' AS date,
    f.capability_statement->>'publisher' AS publisher,
    f.capability_statement->>'description' AS description,
    f.capability_statement->>'purpose' AS purpose,
    f.capability_statement->>'copyright' AS copyright,
    f.capability_statement->'software'->>'name' AS software_name,
    f.capability_statement->'software'->>'version' AS software_version,
    f.capability_statement->'software'->>'releaseDate' AS software_release_date,
    f.capability_statement->'implementation'->>'description' AS implementation_description,
    f.capability_statement->'implementation'->>'url' AS implementation_url,
    f.capability_statement->'implementation'->>'custodian' AS implementation_custodian
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.capability_statement::jsonb != 'null'
AND f.requested_fhir_version = 'None';

-- Create indexes for performance optimization
CREATE INDEX idx_get_capstat_values_mv_endpoint_id ON get_capstat_values_mv(endpoint_id);
CREATE INDEX idx_get_capstat_values_mv_vendor_id ON get_capstat_values_mv(vendor_id);
CREATE INDEX idx_get_capstat_values_mv_filter_fhir_version ON get_capstat_values_mv(filter_fhir_version);
CREATE INDEX idx_get_capstat_values_mv_vendor_name ON get_capstat_values_mv(vendor_name);

-- Create a unique composite index
CREATE UNIQUE INDEX idx_get_capstat_values_mv_unique ON get_capstat_values_mv(endpoint_id, vendor_id, filter_fhir_version);

CREATE MATERIALIZED VIEW get_capstat_fields_mv AS
WITH valid_fhir_versions AS (
    SELECT unnest(ARRAY['No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', 
                         '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', 
                         '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', 
                         '4.0.0', '4.0.1']) AS version
)
SELECT 
    f.id AS endpoint_id,
    f.vendor_id,
    COALESCE(vendors.name, 'Unknown') AS vendor_name,
    CASE 
        -- Extract FHIR version without the part after hyphen
        WHEN POSITION('-' IN f.capability_fhir_version) > 0 
        THEN SUBSTRING(f.capability_fhir_version FROM 1 FOR POSITION('-' IN f.capability_fhir_version) - 1)
        ELSE f.capability_fhir_version 
    END AS raw_version,
    CASE 
        -- Check if simplified version is in the valid_fhir_versions list
        WHEN (
            CASE 
                WHEN POSITION('-' IN f.capability_fhir_version) > 0 
                THEN SUBSTRING(f.capability_fhir_version FROM 1 FOR POSITION('-' IN f.capability_fhir_version) - 1)
                ELSE f.capability_fhir_version 
            END
        ) IN (SELECT version FROM valid_fhir_versions) 
        THEN (
            CASE 
                WHEN POSITION('-' IN f.capability_fhir_version) > 0 
                THEN SUBSTRING(f.capability_fhir_version FROM 1 FOR POSITION('-' IN f.capability_fhir_version) - 1)
                ELSE f.capability_fhir_version 
            END
        ) 
        ELSE 'Unknown' 
    END AS fhir_version,
    json_array_elements(included_fields::json) ->> 'Field' AS field,
    json_array_elements(included_fields::json) ->> 'Exists' AS exist,
    json_array_elements(included_fields::json) ->> 'Extension' AS extension
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE included_fields != 'null' AND requested_fhir_version = 'None'
ORDER BY (json_array_elements(included_fields::json) ->> 'Field');

CREATE UNIQUE INDEX idx_get_capstat_fields_mv_endpoint_id_field ON get_capstat_fields_mv(endpoint_id, field);
CREATE INDEX idx_get_capstat_fields_mv_fhir_version ON get_capstat_fields_mv(fhir_version);
CREATE INDEX idx_get_capstat_fields_mv_field ON get_capstat_fields_mv(field);
CREATE INDEX idx_get_capstat_fields_mv_vendor_id ON get_capstat_fields_mv(vendor_id);

CREATE MATERIALIZED VIEW get_value_versions_mv AS
SELECT 
    field,
    array_agg(DISTINCT fhir_version ORDER BY fhir_version) AS fhir_versions
FROM 
    get_capstat_fields_mv
GROUP BY 
    field;

CREATE UNIQUE INDEX idx_get_value_versions_mv_field ON get_value_versions_mv(field);

CREATE MATERIALIZED VIEW selected_fhir_endpoints_values_mv AS
WITH base_data AS (
    -- Start with the capstat values data
    SELECT 
        g.vendor_name AS "Developer",
        g.filter_fhir_version AS "FHIR Version",
        g.fhir_version AS "fhirVersion",
        g.software_name AS "software.name",
        g.software_version AS "software.version",
        g.software_release_date AS "software.releaseDate",
        g.implementation_description AS "implementation.description",
        g.implementation_url AS "implementation.url",
        g.implementation_custodian AS "implementation.custodian",
        -- All other fields from capability statement
        g.url,
        g.version,
        g.name,
        g.title,
        g.date,
        g.publisher,
        g.description,
        g.purpose,
        g.copyright,
        g.endpoint_id
    FROM get_capstat_values_mv g
),
-- Create a cross join of all possible field combinations
field_combinations AS (
    SELECT 
        b."Developer",
        b."FHIR Version",
        v.field,
        UNNEST(v.fhir_versions) AS field_version,
        -- Create a lateral join to get the value for each field
        CASE 
            WHEN v.field = 'url' THEN b.url
            WHEN v.field = 'version' THEN b.version
            WHEN v.field = 'name' THEN b.name
            WHEN v.field = 'title' THEN b.title
            WHEN v.field = 'date' THEN b.date
            WHEN v.field = 'publisher' THEN b.publisher
            WHEN v.field = 'description' THEN b.description
            WHEN v.field = 'purpose' THEN b.purpose
            WHEN v.field = 'copyright' THEN b.copyright
            WHEN v.field = 'software.name' THEN b."software.name"
            WHEN v.field = 'software.version' THEN b."software.version"
            WHEN v.field = 'software.releaseDate' THEN b."software.releaseDate"
            WHEN v.field = 'implementation.description' THEN b."implementation.description"
            WHEN v.field = 'implementation.url' THEN b."implementation.url"
            WHEN v.field = 'implementation.custodian' THEN b."implementation.custodian"
            ELSE NULL
        END AS field_value,
        b.endpoint_id
    FROM base_data b
    CROSS JOIN get_value_versions_mv v
    WHERE b."FHIR Version" IN (SELECT UNNEST(v.fhir_versions) FROM get_value_versions_mv WHERE field = v.field)
)
-- Final aggregation
SELECT 
    "Developer",
    "FHIR Version",
    field,
    COALESCE(field_value, '[Empty]') AS field_value,
    COUNT(DISTINCT endpoint_id)::INT AS "Endpoints"  -- Explicitly cast to INT
FROM field_combinations
GROUP BY "Developer", "FHIR Version", field, field_value
ORDER BY "Developer", "FHIR Version", field, field_value;

-- Create indexes for performance optimization
CREATE INDEX idx_selected_fhir_endpoints_dev ON selected_fhir_endpoints_values_mv("Developer");
CREATE INDEX idx_selected_fhir_endpoints_fhir_version ON selected_fhir_endpoints_values_mv("FHIR Version");
CREATE INDEX idx_selected_fhir_endpoints_field ON selected_fhir_endpoints_values_mv(Field);
CREATE INDEX idx_selected_fhir_endpoints_field_value ON selected_fhir_endpoints_values_mv(field_value);

-- Create a unique composite index
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_unique ON selected_fhir_endpoints_values_mv("Developer", "FHIR Version", Field, field_value);
