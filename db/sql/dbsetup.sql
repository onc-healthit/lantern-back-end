CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- LANTERN-825: Add history trigger and function
-- Update the history trigger
CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $fhir_endpoints_info_historys$
BEGIN
    -- For INSERT/DELETE operations, always create history
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'D', now(), user, OLD.*;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'I', now(), user, NEW.*;
        RETURN NEW;
    END IF;

    -- For UPDATE operations, check if anything significant changed
    IF (
        NEW.id IS DISTINCT FROM OLD.id OR
        NEW.healthit_mapping_id IS DISTINCT FROM OLD.healthit_mapping_id OR
        NEW.vendor_id IS DISTINCT FROM OLD.vendor_id OR
        NEW.url IS DISTINCT FROM OLD.url OR
        NEW.tls_version IS DISTINCT FROM OLD.tls_version OR
        NEW.mime_types IS DISTINCT FROM OLD.mime_types OR
        NEW.capability_statement::text IS DISTINCT FROM OLD.capability_statement::text OR
        NEW.validation_result_id IS DISTINCT FROM OLD.validation_result_id OR
        NEW.included_fields::text IS DISTINCT FROM OLD.included_fields::text OR
        NEW.operation_resource::text IS DISTINCT FROM OLD.operation_resource::text OR
        NEW.supported_profiles::text IS DISTINCT FROM OLD.supported_profiles::text OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        NEW.smart_response::text IS DISTINCT FROM OLD.smart_response::text OR
        NEW.requested_fhir_version IS DISTINCT FROM OLD.requested_fhir_version OR
        NEW.capability_fhir_version IS DISTINCT FROM OLD.capability_fhir_version
    ) THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'U', now(), user, NEW.*;
    END IF;

    RETURN NEW;
END;
$fhir_endpoints_info_historys$ LANGUAGE plpgsql;

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

-- Add indexes to improve query performance
CREATE INDEX idx_mv_vendor_fhir_counts_vendor ON mv_vendor_fhir_counts(vendor_name);
CREATE INDEX idx_mv_vendor_fhir_counts_fhir ON mv_vendor_fhir_counts(fhir_version);
CREATE INDEX idx_mv_vendor_fhir_counts_sort ON mv_vendor_fhir_counts(sort_order);

-- Create a unique index for concurrent refresh
-- Since vendor_name and fhir_version alone aren't guaranteed to be unique anymore (due to "Others"),
-- we include sort_order to ensure uniqueness
CREATE UNIQUE INDEX idx_mv_vendor_fhir_counts_unique ON mv_vendor_fhir_counts(vendor_name, fhir_version, sort_order);



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
        -- Cast the count to an integer
        COUNT(DISTINCT f.url)::INTEGER AS count_endpoints
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
        -- Cast the sum to an integer to ensure it's not displayed as a decimal
        SUM(count_endpoints)::INTEGER AS count_endpoints
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

-- LANTERN-837
-- endpoint_export_mv
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

-- LANTERN-835

CREATE MATERIALIZED VIEW mv_endpoint_totals AS
WITH latest_metadata AS (
    SELECT max(updated_at) AS last_updated
    FROM fhir_endpoints_metadata
), 
totals AS (
    SELECT 
        -- Count (url, fhir_version) combinations to match Endpoints tab logic
        (SELECT count(*) FROM (SELECT DISTINCT url, fhir_version FROM selected_fhir_endpoints_mv) AS combinations) AS all_endpoints,
        (SELECT count(*) FROM (SELECT DISTINCT fei.url, fei.capability_fhir_version 
        FROM fhir_endpoints_info fei
        WHERE fei.requested_fhir_version = 'None') AS combinations) AS indexed_endpoints
)
SELECT 
    now() AS aggregation_date,
    totals.all_endpoints,
    totals.indexed_endpoints,
    greatest(totals.all_endpoints - totals.indexed_endpoints, 0) AS nonindexed_endpoints,
    (SELECT latest_metadata.last_updated FROM latest_metadata) AS last_updated
FROM totals;

CREATE UNIQUE INDEX idx_mv_endpoint_totals_date ON mv_endpoint_totals(aggregation_date);

-- LANTERN-836: Contacts-MV

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

-- Lantern-856
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

CREATE MATERIALIZED VIEW endpoint_supported_profiles_mv AS
SELECT
  row_number() OVER () AS mv_id,
  f.id AS endpoint_id,
  f.url,
  f.vendor_id,
  COALESCE(vendors.name, 'Unknown') AS vendor_name,
  CASE
    WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
    WHEN split_part(f.capability_fhir_version, '-', 1) = ANY (
      ARRAY[
        'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0',
        '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2',
        '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1'
      ]
    ) THEN split_part(f.capability_fhir_version, '-', 1)
    ELSE 'Unknown'
  END AS fhir_version,
  sp.value ->> 'Resource' AS resource,
  sp.value ->> 'ProfileURL' AS profileurl,
  sp.value ->> 'ProfileName' AS profilename
FROM
  fhir_endpoints_info f
LEFT JOIN
  vendors ON f.vendor_id = vendors.id
CROSS JOIN LATERAL
  json_array_elements(f.supported_profiles::json) sp(value)
WHERE
  f.supported_profiles::text <> 'null'
  AND f.requested_fhir_version = 'None';

CREATE UNIQUE INDEX endpoint_supported_profiles_mv_uidx ON endpoint_supported_profiles_mv(mv_id);
CREATE INDEX idx_profiles_fhir_version ON endpoint_supported_profiles_mv(fhir_version);
CREATE INDEX idx_profiles_vendor_name ON endpoint_supported_profiles_mv(vendor_name);
CREATE INDEX idx_profiles_profileurl ON endpoint_supported_profiles_mv(profileurl);

-- Lantern-852
DROP MATERIALIZED VIEW IF EXISTS mv_capstat_sizes_tbl CASCADE;

CREATE MATERIALIZED VIEW mv_capstat_sizes_tbl AS
SELECT
    f.url,
    pg_column_size(capability_statement::text) AS size,
    CASE
      WHEN REGEXP_REPLACE(
             CASE 
               WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
               ELSE f.capability_fhir_version
             END,
             '-.*', ''
           ) IN (
             'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2',
             '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0',
             '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0',
             '4.0.0', '4.0.1'
           )
      THEN REGEXP_REPLACE(
             CASE 
               WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
               ELSE f.capability_fhir_version
             END,
             '-.*', ''
           )
      ELSE 'Unknown'
    END AS fhir_version,
    COALESCE(vendors.name, 'Unknown') AS vendor_name
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.capability_fhir_version != ''
  AND f.requested_fhir_version = 'None';

-- Create indexes for mv_capstat_sizes
CREATE UNIQUE INDEX idx_mv_capstat_sizes_uniq ON mv_capstat_sizes_tbl(url);
CREATE INDEX idx_mv_capstat_sizes_fhir ON mv_capstat_sizes_tbl(fhir_version);
CREATE INDEX idx_mv_capstat_sizes_vendor ON mv_capstat_sizes_tbl(vendor_name);


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
	    WHEN f.capability_fhir_version LIKE '%-%' THEN SPLIT_PART(f.capability_fhir_version, '-', 1)
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
            WHEN v.field = 'fhirVersion' THEN b."fhirVersion"
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
    CASE WHEN COALESCE(field_value, '[Empty]') = '[Empty]' THEN 'no' ELSE 'yes' END AS is_used,
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
CREATE INDEX idx_selected_fhir_endpoints_is_used ON selected_fhir_endpoints_values_mv(is_used);
CREATE INDEX idx_summary_query ON selected_fhir_endpoints_values_mv (field, "FHIR Version", "Developer", is_used);

-- Create a unique composite index
CREATE UNIQUE INDEX idx_selected_fhir_endpoints_unique ON selected_fhir_endpoints_values_mv("Developer", "FHIR Version", Field, field_value);

-- Add capstat_usage_summary_mv
CREATE MATERIALIZED VIEW capstat_usage_summary_mv AS
SELECT 
  field,
  "FHIR Version",
  "Developer",
  is_used,
  SUM("Endpoints") AS count
FROM selected_fhir_endpoints_values_mv
GROUP BY field, "FHIR Version", "Developer", is_used;

CREATE INDEX idx_usage_summary_filters ON capstat_usage_summary_mv(field, "FHIR Version", "Developer", is_used);

CREATE TABLE daily_querying_status (status VARCHAR(500));
INSERT INTO daily_querying_status VALUES ('true');

-- Lantern-839
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

WITH DATA;

 -- Create indexes for endpoint list organizations materialized view
 CREATE UNIQUE INDEX idx_mv_endpoint_list_org_uniq ON mv_endpoint_list_organizations(fhir_version, vendor_name, url, organization_name, organization_id);
 CREATE INDEX idx_mv_endpoint_list_org_fhir ON mv_endpoint_list_organizations(fhir_version);
 CREATE INDEX idx_mv_endpoint_list_org_vendor ON mv_endpoint_list_organizations(vendor_name);
 CREATE INDEX idx_mv_endpoint_list_org_url ON mv_endpoint_list_organizations(url);

CREATE MATERIALIZED VIEW mv_validation_results_plot AS
SELECT DISTINCT t.url,
t.fhir_version,
t.vendor_name,
t.rule_name,
t.valid,
t.expected,
t.actual,
t.comment,
t.reference
FROM ( SELECT COALESCE(vendors.name, 'Unknown'::character varying) AS vendor_name,
        f.url,
            CASE
                WHEN f.capability_fhir_version::text = ''::text THEN 'No Cap Stat'::character varying
                WHEN "position"(f.capability_fhir_version::text, '-'::text) > 0 THEN "substring"(f.capability_fhir_version::text, 1, "position"(f.capability_fhir_version::text, '-'::text) - 1)::character varying
                WHEN f.capability_fhir_version::text <> ALL (ARRAY['0.4.0'::character varying, '0.5.0'::character varying, '1.0.0'::character varying, '1.0.1'::character varying, '1.0.2'::character varying, '1.1.0'::character varying, '1.2.0'::character varying, '1.4.0'::character varying, '1.6.0'::character varying, '1.8.0'::character varying, '3.0.0'::character varying, '3.0.1'::character varying, '3.0.2'::character varying, '3.2.0'::character varying, '3.3.0'::character varying, '3.5.0'::character varying, '3.5a.0'::character varying, '4.0.0'::character varying, '4.0.1'::character varying]::text[]) THEN 'Unknown'::character varying
                ELSE f.capability_fhir_version
            END AS fhir_version,
        v.rule_name,
        v.valid,
        v.expected,
        v.actual,
        v.comment,
        v.reference,
        v.validation_result_id AS id,
        f.requested_fhir_version
        FROM fhir_endpoints_info f
            JOIN validations v ON f.validation_result_id = v.validation_result_id
            LEFT JOIN vendors ON f.vendor_id = vendors.id
        ORDER BY v.validation_result_id, v.rule_name) t;

CREATE UNIQUE INDEX mv_validation_results_plot_unique_idx 
ON mv_validation_results_plot(url, fhir_version, vendor_name, rule_name, valid, expected, actual);

CREATE INDEX mv_validation_results_plot_vendor_idx ON mv_validation_results_plot(vendor_name);
CREATE INDEX mv_validation_results_plot_fhir_idx ON mv_validation_results_plot(fhir_version);
CREATE INDEX mv_validation_results_plot_rule_idx ON mv_validation_results_plot(rule_name);
CREATE INDEX mv_validation_results_plot_valid_idx ON mv_validation_results_plot(valid);
CREATE INDEX mv_validation_results_plot_reference_idx ON mv_validation_results_plot(reference);

-- Materialized view for validation details
CREATE MATERIALIZED VIEW mv_validation_details AS 
WITH validation_data AS ( 
    SELECT 
        COALESCE(vendors.name, 'Unknown') as vendor_name, 
        CASE 
            WHEN capability_fhir_version = '' THEN 'No Cap Stat' 
            WHEN position('-' in capability_fhir_version) > 0 THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1) 
            WHEN capability_fhir_version NOT IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'Unknown' 
            ELSE capability_fhir_version 
        END AS fhir_version, 
        rule_name, 
        reference 
    FROM validations v 
    JOIN fhir_endpoints_info f ON v.validation_result_id = f.validation_result_id 
    LEFT JOIN vendors on f.vendor_id = vendors.id 
    WHERE f.requested_fhir_version = 'None'
    AND v.rule_name IS NOT NULL 
),
mapped_versions AS (
    SELECT DISTINCT
        rule_name,
        CASE 
            WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2' 
            WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3' 
            WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4' 
            ELSE fhir_version
        END AS version_name,
        -- Add a sort order to maintain the original ordering
        CASE
            WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 1 -- DSTU2
            WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 2 -- STU3
            WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 3 -- R4
            ELSE 4 -- Others
        END AS sort_order
    FROM validation_data
    WHERE fhir_version != 'Unknown' AND fhir_version != 'No Cap Stat'
),
validation_versions AS (
    SELECT
        rule_name,
        STRING_AGG(version_name, ', ' ORDER BY sort_order) as fhir_version_names
    FROM (
        SELECT DISTINCT rule_name, version_name, sort_order
        FROM mapped_versions
    ) AS unique_versions
    GROUP BY rule_name
)
SELECT 
    vd.rule_name, 
    COALESCE(vv.fhir_version_names, '') as fhir_version_names 
FROM ( 
    SELECT DISTINCT rule_name 
    FROM validation_data 
) vd 
LEFT JOIN validation_versions vv ON vd.rule_name = vv.rule_name 
ORDER BY vd.rule_name;

CREATE UNIQUE INDEX mv_validation_details_unique_idx ON mv_validation_details(rule_name); 

-- Materialized view for validation failures
CREATE MATERIALIZED VIEW mv_validation_failures AS
SELECT fhir_version, url, expected, actual, vendor_name, rule_name, reference
FROM mv_validation_results_plot
WHERE valid = 'false';

CREATE UNIQUE INDEX mv_validation_failures_unique_idx ON mv_validation_failures(url, fhir_version, vendor_name, rule_name);
CREATE INDEX mv_validation_failures_url_idx ON mv_validation_failures(url);
CREATE INDEX mv_validation_failures_fhir_version_idx ON mv_validation_failures(fhir_version);
CREATE INDEX mv_validation_failures_vendor_name_idx ON mv_validation_failures(vendor_name);
CREATE INDEX mv_validation_failures_rule_name_idx ON mv_validation_failures(rule_name);
CREATE INDEX mv_validation_failures_reference_idx ON mv_validation_failures(reference);

--LANTERN-security_tab_mv
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

-- LANTERN-843
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

-- Create materialized view for security_endpoints_distinct_mv
DROP MATERIALIZED VIEW IF EXISTS security_endpoints_distinct_mv CASCADE;
CREATE MATERIALIZED VIEW security_endpoints_distinct_mv AS
SELECT DISTINCT
  url_modal AS url,
  condensed_organization_names,
  vendor_name,
  capability_fhir_version,
  tls_version,
  code
FROM selected_security_endpoints_mv;

-- Create indexes for security_endpoints_distinct_mv
CREATE UNIQUE INDEX idx_unique_security_endpoints_distinct_mv ON security_endpoints_distinct_mv (url, condensed_organization_names, vendor_name, capability_fhir_version, tls_version, code);
CREATE INDEX idx_security_endpoints_distinct_filters  ON security_endpoints_distinct_mv(capability_fhir_version, code, vendor_name);

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

-- Lantern-854
-- Create materialized view for capstat_fields
CREATE MATERIALIZED VIEW mv_capstat_fields AS 

SELECT 
  f.id AS endpoint_id,
  f.vendor_id,
  COALESCE(vendors.name, 'Unknown') AS vendor_name,
  CASE
    WHEN REGEXP_REPLACE(f.capability_fhir_version, '-.*', '') IN (
      'No Cap Stat', '0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2',
      '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0',
      '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0',
      '4.0.0', '4.0.1'
    )
    THEN REGEXP_REPLACE(f.capability_fhir_version, '-.*', '')
    ELSE 'Unknown'
  END AS fhir_version,
  json_array_elements(f.included_fields::json) ->> 'Field' AS field,
  json_array_elements(f.included_fields::json) ->> 'Exists' AS exist,
  json_array_elements(f.included_fields::json) ->> 'Extension' AS extension
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.included_fields != 'null'
  AND f.requested_fhir_version = 'None'
ORDER BY endpoint_id;

-- Create indexes for mv_capstat_fields
CREATE UNIQUE INDEX idx_mv_capstat_fields_unique ON mv_capstat_fields(endpoint_id, fhir_version, field);
CREATE INDEX idx_mv_capstat_fields_vendor ON mv_capstat_fields(vendor_name);
CREATE INDEX idx_mv_capstat_fields_fhir ON mv_capstat_fields(fhir_version);

-- Create materialized view for capstat_fields_text
CREATE MATERIALIZED VIEW mv_capstat_values_fields AS 

WITH base AS (
  -- Start with the materialized view filtered on extension
  SELECT *
  FROM mv_capstat_fields
  WHERE extension = 'false'
),
with_version AS (
  -- Add a computed column for the FHIR version name based on your groups
  SELECT
    field,
    fhir_version,
    CASE
      WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2'
      WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3'
      WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4'
      ELSE 'DSTU2'
    END AS fhir_version_name
  FROM base
),
fvn AS (
  -- For each field, get the unique FHIR version names as a comma‐separated list
  SELECT 
    field, 
    string_agg(DISTINCT fhir_version_name, ', ') AS fhir_version_names
  FROM with_version
  GROUP BY field
)
SELECT DISTINCT 
  fhir_version,
  with_version.field || ' (' || fvn.fhir_version_names || ')' AS field_version
FROM with_version
LEFT JOIN fvn ON with_version.field = fvn.field
ORDER BY field_version;

-- Create indexes for mv_capstat_values_fields
CREATE UNIQUE INDEX idx_mv_capstat_values_fields_unique ON mv_capstat_values_fields(fhir_version, field_version);
CREATE INDEX idx_mv_capstat_values_fields_field_version ON mv_capstat_values_fields(field_version);
CREATE INDEX idx_mv_capstat_values_fields_fhir ON mv_capstat_values_fields(fhir_version);

-- Create materialized view for capstat_extension_text
CREATE MATERIALIZED VIEW mv_capstat_values_extension AS 

WITH base AS (
  -- Start with the materialized view filtered on extension
  SELECT *
  FROM mv_capstat_fields
  WHERE extension = 'true'
),
with_version AS (
  -- Add a computed column for the FHIR version name based on your groups
  SELECT
    field,
    fhir_version,
    CASE
      WHEN fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2') THEN 'DSTU2'
      WHEN fhir_version IN ('1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2') THEN 'STU3'
      WHEN fhir_version IN ('3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') THEN 'R4'
      ELSE 'DSTU2'
    END AS fhir_version_name
  FROM base
),
fvn AS (
  -- For each field, get the unique FHIR version names as a comma‐separated list
  SELECT 
    field, 
    string_agg(DISTINCT fhir_version_name, ', ') AS fhir_version_names
  FROM with_version
  GROUP BY field
)
SELECT DISTINCT 
  fhir_version,
  with_version.field || ' (' || fvn.fhir_version_names || ')' AS field_version
FROM with_version
LEFT JOIN fvn ON with_version.field = fvn.field
ORDER BY field_version;

-- Create indexes for mv_capstat_values_extension
CREATE UNIQUE INDEX idx_mv_capstat_values_extension_unique ON mv_capstat_values_extension(fhir_version, field_version);
CREATE INDEX idx_mv_capstat_values_extension_field_version ON mv_capstat_values_extension(field_version);
CREATE INDEX idx_mv_capstat_values_extension_fhir ON mv_capstat_values_extension(fhir_version);

-- LANTERN-863
-- Create materialized view for removing resource fetcher

CREATE MATERIALIZED VIEW mv_endpoint_resource_types AS
SELECT 
    f.id AS endpoint_id,
    f.vendor_id,
    COALESCE(vendors.name, 'Unknown') AS vendor_name,
    CASE 
        WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
        WHEN position('-' in f.capability_fhir_version) > 0 THEN substring(f.capability_fhir_version from 1 for position('-' in f.capability_fhir_version) - 1)
        WHEN f.capability_fhir_version IN ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1') 
            THEN f.capability_fhir_version
        ELSE 'Unknown'
    END AS fhir_version,
    json_array_elements(capability_statement::json#>'{rest,0,resource}') ->> 'type' AS type
FROM fhir_endpoints_info f
LEFT JOIN vendors ON f.vendor_id = vendors.id
WHERE f.requested_fhir_version = 'None'
ORDER BY type;

-- Create indexes for better performance
CREATE UNIQUE INDEX idx_mv_endpoint_resource_types_unique ON mv_endpoint_resource_types(endpoint_id, vendor_id, fhir_version, type);
CREATE INDEX idx_mv_endpoint_resource_types_vendor ON mv_endpoint_resource_types(vendor_name);
CREATE INDEX idx_mv_endpoint_resource_types_fhir ON mv_endpoint_resource_types(fhir_version);
CREATE INDEX idx_mv_endpoint_resource_types_type ON mv_endpoint_resource_types(type);

-- LANTERN-864
CREATE MATERIALIZED VIEW mv_get_security_endpoints AS
SELECT
  f.id,
  f.vendor_id,
  COALESCE(v.name, 'Unknown') AS name,
  CASE 
    WHEN capability_fhir_version = '' THEN 'No Cap Stat'
    WHEN position('-' in capability_fhir_version) > 0 THEN 
      CASE
        WHEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1) IN 
            ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
             '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
        THEN substring(capability_fhir_version, 1, position('-' in capability_fhir_version) - 1)
        ELSE 'Unknown'
      END
    WHEN capability_fhir_version IN 
        ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
         '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
    THEN capability_fhir_version
    ELSE 'Unknown'
  END AS fhir_version,
  json_array_elements(json_array_elements(capability_statement::json#>'{rest,0,security,service}')->'coding')::json->>'code' AS code,
  json_array_elements(capability_statement::json#>'{rest,0,security}' -> 'service')::json ->> 'text' AS text
FROM fhir_endpoints_info f 
LEFT JOIN vendors v ON f.vendor_id = v.id
WHERE requested_fhir_version = 'None';

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_get_security_endpoints ON mv_get_security_endpoints(id, code);
CREATE INDEX idx_mv_get_security_endpoints_name ON mv_get_security_endpoints(name);
CREATE INDEX idx_mv_get_security_endpoints_fhir ON mv_get_security_endpoints(fhir_version);

CREATE MATERIALIZED VIEW mv_auth_type_count AS
WITH endpoints_by_version AS (
  -- Get total count of distinct IDs per FHIR version
  SELECT 
    fhir_version,
    COUNT(DISTINCT id) AS tc
  FROM 
    mv_get_security_endpoints
  GROUP BY 
    fhir_version
),
endpoints_by_version_code AS (
  -- Count endpoints for each code within each FHIR version
  SELECT 
    s.fhir_version,
    s.code,
    e.tc,
    COUNT(DISTINCT s.id) AS endpoints
  FROM 
    mv_get_security_endpoints s
  JOIN 
    endpoints_by_version e ON s.fhir_version = e.fhir_version
  GROUP BY 
    s.fhir_version, s.code, e.tc
)
-- Calculate final results with percentages
SELECT 
  code AS "Code",
  fhir_version AS "FHIR Version",
  endpoints::integer AS "Endpoints",
  ROUND(endpoints::numeric * 100 / tc)::integer || '%' AS "Percent"
FROM 
  endpoints_by_version_code
ORDER BY 
  "FHIR Version",  
  "Code"; 

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_auth_type_count ON mv_auth_type_count("Code", "FHIR Version");
CREATE INDEX idx_mv_auth_type_count_fhir ON mv_auth_type_count("FHIR Version");
CREATE INDEX idx_mv_auth_type_count_endpoints ON mv_auth_type_count("Endpoints"); 

CREATE MATERIALIZED VIEW mv_endpoint_security_counts AS
WITH 
-- Get total indexed endpoints from mv_endpoint_totals
total_endpoints AS (
  SELECT 
    'Total Indexed Endpoints' AS status,
    all_endpoints::integer AS endpoints,
    1 AS sort_order
  FROM mv_endpoint_totals
  ORDER BY aggregation_date DESC
  LIMIT 1
),
-- Get HTTP 200 responses from mv_response_tally
http_200_endpoints AS (
  SELECT 
    'Endpoints with successful response (HTTP 200)' AS status,
    http_200::integer AS endpoints,
    2 AS sort_order
  FROM mv_response_tally
  LIMIT 1
),
-- Get non-200 responses from mv_response_tally
http_non200_endpoints AS (
  SELECT 
    'Endpoints with unsuccessful response' AS status,
    http_non200::integer AS endpoints,
    3 AS sort_order
  FROM mv_response_tally
  LIMIT 1
),
-- Get count of endpoints without valid capability statement
no_cap_statement AS (
  SELECT 
    'Endpoints without valid CapabilityStatement / Conformance Resource' AS status,
    COUNT(*)::integer AS endpoints,
    4 AS sort_order
  FROM fhir_endpoints_info 
  WHERE jsonb_typeof(capability_statement::jsonb) <> 'object' 
    AND requested_fhir_version = 'None'
),
-- Get count of endpoints with valid security resource
security_endpoints AS (
  SELECT 
    'Endpoints with valid security resource' AS status,
    COUNT(DISTINCT id)::integer AS endpoints,
    5 AS sort_order
  FROM mv_get_security_endpoints
),
-- Combine all results
combined_results AS (
  SELECT status, endpoints, sort_order FROM total_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM http_200_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM http_non200_endpoints
  UNION ALL
  SELECT status, endpoints, sort_order FROM no_cap_statement
  UNION ALL
  SELECT status, endpoints, sort_order FROM security_endpoints
)
-- Final select with ordering
SELECT 
  status AS "Status",
  endpoints AS "Endpoints"
FROM combined_results
ORDER BY sort_order;

-- Create a unique index
CREATE UNIQUE INDEX idx_mv_endpoint_security_counts ON mv_endpoint_security_counts("Status");

-- LANTERN-838: Validation cleanup 
CREATE INDEX fhir_endpoints_info_history_val_res_idx ON fhir_endpoints_info_history (validation_result_id);

ALTER TABLE validations
ADD CONSTRAINT fk_validations_validation_results
FOREIGN KEY (validation_result_id) 
REFERENCES validation_results(id)
ON DELETE CASCADE;

-- LANTERN-841: HTI-1 Final Rule Organization Data
CREATE TABLE fhir_endpoint_organization_active (
	org_id INT,
	active VARCHAR(500)
);

CREATE TABLE fhir_endpoint_organization_addresses (
	org_id INT,
	address VARCHAR(500)
);

CREATE TABLE fhir_endpoint_organization_identifiers (
	org_id INT,
	identifier VARCHAR(500)
);

CREATE INDEX idx_fhir_endpoint_organization_active_org_id ON fhir_endpoint_organization_active (org_id);

CREATE INDEX idx_fhir_endpoint_organization_addresses_org_id ON fhir_endpoint_organization_addresses (org_id);

CREATE INDEX idx_fhir_endpoint_organization_identifiers_org_id ON fhir_endpoint_organization_identifiers (org_id);

CREATE TABLE fhir_endpoint_organization_url (
	org_id INT,
	org_url VARCHAR(500)
);

CREATE INDEX idx_fhir_endpoint_organization_url_org_id ON fhir_endpoint_organization_url (org_id);

--Profiles Tab Pagination MV
CREATE MATERIALIZED VIEW mv_profiles_paginated AS
SELECT
  row_number() OVER (ORDER BY vendor_name, url, profileurl) AS page_id,
  url,
  profileurl,
  profilename,
  resource,
  fhir_version,
  vendor_name
FROM (
  SELECT DISTINCT
    url,
    profileurl,
    profilename,
    resource,
    fhir_version,
    vendor_name
  FROM endpoint_supported_profiles_mv
) distinct_profiles
ORDER BY vendor_name, url, profileurl;

-- Create indexes for fast filtering and pagination
CREATE UNIQUE INDEX mv_profiles_paginated_page_id_idx ON mv_profiles_paginated(page_id);
CREATE INDEX mv_profiles_paginated_fhir_version_idx ON mv_profiles_paginated(fhir_version);
CREATE INDEX mv_profiles_paginated_vendor_name_idx ON mv_profiles_paginated(vendor_name);
CREATE INDEX mv_profiles_paginated_resource_idx ON mv_profiles_paginated(resource);
CREATE INDEX mv_profiles_paginated_profileurl_idx ON mv_profiles_paginated(profileurl);

-- Composite index for common filter combinations
CREATE INDEX mv_profiles_paginated_composite_idx ON mv_profiles_paginated(vendor_name, fhir_version, resource);

-- LANTERN-925: Improve pagination performance of the organization tab
CREATE MATERIALIZED VIEW mv_organizations_aggregated AS
WITH base_filtered_data AS (
    -- Step 1: Get the source data from mv_endpoint_list_organizations
    SELECT 
        mv.organization_name,
        mv.organization_id,
        mv.url,
        mv.fhir_version,
        mv.vendor_name
    FROM mv_endpoint_list_organizations mv
),
processed_data AS (
    -- Step 2: Apply the R mutate logic including uppercase conversion
    SELECT DISTINCT
        -- Replicate tidyr::replace_na(list(organization_name = "Unknown")) + UPPER()
        CASE 
            WHEN organization_name IS NULL OR organization_name = '' THEN 'UNKNOWN'
            ELSE UPPER(organization_name)
        END AS organization_name,
        -- Replicate mutate(organization_id = as.integer(organization_id))
        CASE 
            WHEN organization_id IS NULL OR organization_id = '' OR organization_id = 'Unknown' THEN NULL
            WHEN organization_id ~ '^[0-9]+$' THEN organization_id::integer
            ELSE NULL
        END as org_id,
        url,
        -- Replicate the consistent FHIR version processing
        CASE 
            WHEN fhir_version = '' OR fhir_version IS NULL THEN 'No Cap Stat'
            WHEN position('-' in fhir_version) > 0 THEN 
                CASE
                    WHEN substring(fhir_version, 1, position('-' in fhir_version) - 1) IN 
                        ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
                         '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
                    THEN substring(fhir_version, 1, position('-' in fhir_version) - 1)
                    ELSE 'Unknown'
                END
            WHEN fhir_version IN 
                ('0.4.0', '0.5.0', '1.0.0', '1.0.1', '1.0.2', '1.1.0', '1.2.0', '1.4.0', '1.6.0', '1.8.0', 
                 '3.0.0', '3.0.1', '3.0.2', '3.2.0', '3.3.0', '3.5.0', '3.5a.0', '4.0.0', '4.0.1', 'No Cap Stat')
            THEN fhir_version
            ELSE 'Unknown'
        END AS fhir_version,
        vendor_name
    FROM base_filtered_data
    WHERE organization_id IS NOT NULL AND organization_id != '' AND organization_id != 'Unknown'
),
-- Step 3: Get distinct org ID and split identifiers into type and value
identifiers_raw AS (
    SELECT DISTINCT
        pd.org_id,
        -- Split identifier into type and value parts
        CASE 
            WHEN fei.identifier ~ '^[^:]+: ' THEN 
                TRIM(substring(fei.identifier from '^([^:]+):'))
            ELSE 'Unknown'
        END as identifier_type,
        CASE 
            WHEN fei.identifier ~ '^[^:]+: ' THEN 
                TRIM(substring(fei.identifier from '^[^:]+: (.*)$'))
            ELSE fei.identifier
        END as identifier_value
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_identifiers fei ON pd.org_id = fei.org_id
    WHERE fei.identifier IS NOT NULL
),
identifiers_agg AS (
    SELECT 
        org_id,
        -- HTML format for identifier types
        string_agg(identifier_type, '<br/>' ORDER BY identifier_type, identifier_value) as identifier_types_html,
        -- HTML format for identifier values
        string_agg(identifier_value, '<br/>' ORDER BY identifier_type, identifier_value) as identifier_values_html,
        -- CSV format for identifier types
        CASE 
            WHEN LENGTH(string_agg(identifier_type, E'\n' ORDER BY identifier_type, identifier_value)) <= 32765 
            THEN string_agg(identifier_type, E'\n' ORDER BY identifier_type, identifier_value)
            ELSE 
                LEFT(
                    string_agg(identifier_type, E'\n' ORDER BY identifier_type, identifier_value), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier_type, E'\n' ORDER BY identifier_type, identifier_value), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier_type, E'\n' ORDER BY identifier_type, identifier_value), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as identifier_types_csv,
        -- CSV format for identifier values (maintain order and duplicates, truncate at complete lines)
        CASE 
            WHEN LENGTH(string_agg(identifier_value, E'\n' ORDER BY identifier_type, identifier_value)) <= 32765 
            THEN string_agg(identifier_value, E'\n' ORDER BY identifier_type, identifier_value)
            ELSE 
                LEFT(
                    string_agg(identifier_value, E'\n' ORDER BY identifier_type, identifier_value), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier_value, E'\n' ORDER BY identifier_type, identifier_value), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(identifier_value, E'\n' ORDER BY identifier_type, identifier_value), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as identifier_values_csv
    FROM identifiers_raw
    GROUP BY org_id
),
-- Step 4: Get DISTINCT addresses per organization ID
addresses_raw AS (
    SELECT DISTINCT
        pd.org_id,
        UPPER(fea.address) as address
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_addresses fea ON pd.org_id = fea.org_id
    WHERE fea.address IS NOT NULL
),
addresses_agg AS (
    SELECT 
        org_id,
        string_agg(address, '<br/>') as addresses_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(address, E'\n')) <= 32765 
            THEN string_agg(address, E'\n')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(address, E'\n'), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(address, E'\n'), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(address, E'\n'), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as addresses_csv
    FROM addresses_raw
    GROUP BY org_id
),
-- Step 5: Get DISTINCT org URLs per organization ID with urn:uuid filtering
urls_raw AS (
    SELECT DISTINCT
        pd.org_id,
        -- Apply the urn:uuid filtering 
        CASE 
            WHEN feou.org_url LIKE 'urn:uuid:%' THEN ''
            ELSE feou.org_url
        END as org_url
    FROM processed_data pd
    LEFT JOIN fhir_endpoint_organization_url feou ON pd.org_id = feou.org_id
    WHERE feou.org_url IS NOT NULL AND feou.org_url != ''
),
urls_agg AS (
    SELECT 
        org_id,
        string_agg(org_url, '<br/>') FILTER (WHERE org_url != '') as org_urls_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(org_url, E'\n') FILTER (WHERE org_url != '')) <= 32765 
            THEN string_agg(org_url, E'\n') FILTER (WHERE org_url != '')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(org_url, E'\n') FILTER (WHERE org_url != ''), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as org_urls_csv
    FROM urls_raw
    GROUP BY org_id
),
-- Step 6: Group by organization ID 
endpoint_data_agg AS (
    SELECT 
        org_id,
        -- Use any organization name for this org_id (they should all be the same after UPPER conversion)
        MAX(organization_name) as organization_name,
        -- HTML formatted endpoint URLs
        string_agg(
            DISTINCT '<a class="lantern-url" tabindex="0" aria-label="Press enter to open a pop up modal containing additional information for this endpoint." onkeydown="javascript:(function(event) { if (event.keyCode === 13){event.target.click()}})(event)" onclick="Shiny.setInputValue(''endpoint_popup'',&quot;' || url || '&quot,{priority: ''event''});"> ' || url || '</a>',
            '<br/>'
        ) as endpoint_urls_html,
        -- Truncate at complete lines to prevent CSV corruption
        CASE 
            WHEN LENGTH(string_agg(DISTINCT url, E'\n')) <= 32765 
            THEN string_agg(DISTINCT url, E'\n')
            ELSE 
                -- Find the last complete line within 32765 chars
                LEFT(
                    string_agg(DISTINCT url, E'\n'), 
                    GREATEST(
                        1,
                        CASE 
                            WHEN POSITION(E'\n' IN REVERSE(LEFT(string_agg(DISTINCT url, E'\n'), 32765))) > 0
                            THEN 32765 - POSITION(E'\n' IN REVERSE(LEFT(string_agg(DISTINCT url, E'\n'), 32765))) + 1
                            ELSE 32765
                        END
                    )
                )
        END as endpoint_urls_csv,
        string_agg(DISTINCT fhir_version, '<br/>') as fhir_versions_html,
        string_agg(DISTINCT fhir_version, E'\n') as fhir_versions_csv,
        string_agg(DISTINCT vendor_name, '<br/>') as vendor_names_html,
        string_agg(DISTINCT vendor_name, E'\n') as vendor_names_csv,
        -- Arrays for filtering (exactly as original code)
        ARRAY(SELECT DISTINCT unnest(array_agg(fhir_version))::text ORDER BY unnest)::text[] as fhir_versions_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(vendor_name))::text ORDER BY unnest)::text[] as vendor_names_array,
        ARRAY(SELECT DISTINCT unnest(array_agg(url))::text ORDER BY unnest)::text[] as urls_array
    FROM processed_data
    GROUP BY org_id  -- KEY CHANGE: Group by org_id instead of organization_name
)
-- Step 7: Final select with JOINs to get all related data per organization ID
SELECT 
    eda.organization_name,
    eda.org_id,
    -- For HTML display (pagination) - split identifier columns
    COALESCE(ia.identifier_types_html, '') as identifier_types_html,
    COALESCE(ia.identifier_values_html, '') as identifier_values_html,
    COALESCE(aa.addresses_html, '') as addresses_html,
    eda.endpoint_urls_html,
    COALESCE(ua.org_urls_html, '') as org_urls_html,
    eda.fhir_versions_html,
    eda.vendor_names_html,
    
    -- For CSV export - split identifier columns
    COALESCE(ia.identifier_types_csv, '') as identifier_types_csv,
    COALESCE(ia.identifier_values_csv, '') as identifier_values_csv,
    COALESCE(aa.addresses_csv, '') as addresses_csv,
    eda.endpoint_urls_csv,
    COALESCE(ua.org_urls_csv, '') as org_urls_csv,
    eda.fhir_versions_csv,
    eda.vendor_names_csv,
    
    -- Arrays for filtering 
    eda.fhir_versions_array,
    eda.vendor_names_array,
    eda.urls_array
    
FROM endpoint_data_agg eda
LEFT JOIN identifiers_agg ia ON eda.org_id = ia.org_id
LEFT JOIN addresses_agg aa ON eda.org_id = aa.org_id  
LEFT JOIN urls_agg ua ON eda.org_id = ua.org_id
WHERE eda.organization_name != 'UNKNOWN'
ORDER BY eda.organization_name, eda.org_id;

-- Create indexes for performance 
CREATE UNIQUE INDEX idx_mv_orgs_agg_org_id ON mv_organizations_aggregated(org_id);
CREATE INDEX idx_mv_orgs_agg_name ON mv_organizations_aggregated(organization_name);
CREATE INDEX idx_mv_orgs_agg_fhir_versions ON mv_organizations_aggregated USING GIN(fhir_versions_array);
CREATE INDEX idx_mv_orgs_agg_vendor_names ON mv_organizations_aggregated USING GIN(vendor_names_array);
CREATE INDEX idx_mv_orgs_agg_urls ON mv_organizations_aggregated USING GIN(urls_array);

--LANTERN-973: Group organizations in Org Table if all the fields are same except endpoint url
CREATE MATERIALIZED VIEW mv_organizations_final AS
SELECT 
    ROW_NUMBER() OVER (ORDER BY organization_name) as org_id,
    organization_name,
    identifier_types_html,
    identifier_values_html,
    addresses_html,
    org_urls_html,
    -- Combine endpoint URLs where everything else matches
    string_agg(DISTINCT endpoint_urls_html, '<br/>') as endpoint_urls_html,
    fhir_versions_html,
    vendor_names_html,
    
    -- CSV versions
    identifier_types_csv,
    identifier_values_csv,
    addresses_csv,
    org_urls_csv,
    string_agg(DISTINCT endpoint_urls_csv, E'\n') as endpoint_urls_csv,
    fhir_versions_csv,
    vendor_names_csv,
    
    -- Arrays for filtering (combine from all matching rows) - FIXED: Use |||| delimiter instead of comma
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(fhir_versions_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as fhir_versions_array,
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(vendor_names_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as vendor_names_array,
    ARRAY(SELECT DISTINCT elem FROM unnest(string_to_array(string_agg(array_to_string(urls_array, '||||'), '||||'), '||||')) AS elem ORDER BY elem) as urls_array
    
FROM mv_organizations_aggregated
GROUP BY 
    organization_name,
    identifier_types_html,
    identifier_values_html,
    addresses_html,
    org_urls_html,
    fhir_versions_html,
    vendor_names_html,
    identifier_types_csv,
    identifier_values_csv,
    addresses_csv,
    org_urls_csv,
    fhir_versions_csv,
    vendor_names_csv
ORDER BY organization_name;

-- Create indexes for performance
CREATE UNIQUE INDEX idx_mv_orgs_final_org_id ON mv_organizations_final(org_id);
CREATE INDEX idx_mv_orgs_final_name ON mv_organizations_final(organization_name);
CREATE INDEX idx_mv_orgs_final_fhir_versions ON mv_organizations_final USING GIN(fhir_versions_array);
CREATE INDEX idx_mv_orgs_final_vendor_names ON mv_organizations_final USING GIN(vendor_names_array);
CREATE INDEX idx_mv_orgs_final_urls ON mv_organizations_final USING GIN(urls_array);