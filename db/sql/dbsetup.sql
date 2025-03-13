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

-- LANTERN-836: Create an SQL Materialized View for the Contact Information Tab
CREATE MATERIALIZED VIEW mv_contact_information AS
WITH contact_data AS (
  -- Get contact information from JSON
  SELECT
    f.url,
    f.requested_fhir_version,
    COALESCE(v.name, 'Unknown') AS vendor_name,
    CASE WHEN f.capability_fhir_version = '' THEN 'No Cap Stat'
         WHEN f.capability_fhir_version SIMILAR TO '[0-9]+\.[0-9]+\.[0-9]+-.*' THEN SUBSTRING(f.capability_fhir_version FROM 1 FOR POSITION('-' IN f.capability_fhir_version)-1)
         ELSE f.capability_fhir_version
    END AS fhir_version,
    e.endpoint_names,
    contact_obj->>'name' AS contact_name,
    telecom_obj->>'system' AS contact_type,
    telecom_obj->>'value' AS contact_value,
    COALESCE((telecom_obj->>'rank')::integer, 999) AS contact_preference
  FROM fhir_endpoints_info f
  LEFT JOIN vendors v ON f.vendor_id = v.id
  LEFT JOIN endpoint_export e ON f.url = e.url AND f.requested_fhir_version = e.requested_fhir_version
  LEFT JOIN LATERAL jsonb_array_elements(f.capability_statement::jsonb->'contact') contact_obj
    ON f.capability_statement::jsonb != 'null'
  LEFT JOIN LATERAL jsonb_array_elements(contact_obj->'telecom') telecom_obj
    ON TRUE
  WHERE f.requested_fhir_version = 'None'
),
endpoints_with_metrics AS (
  -- Calculate metrics and prepare for final view
  SELECT
    cd.url,
    cd.requested_fhir_version,
    cd.vendor_name,
    cd.fhir_version,
    cd.endpoint_names,
    cd.contact_name,
    cd.contact_type,
    cd.contact_value,
    cd.contact_preference,
    -- Pre-process endpoint names for display (handling as text)
    -- Pre-process endpoint names for display (handling as text and removing braces/quotes)
CASE 
  WHEN cd.endpoint_names IS NULL THEN NULL
  WHEN cd.endpoint_names::text = '' THEN NULL
  ELSE
    -- Remove curly braces and quotes
    REGEXP_REPLACE(
      REGEXP_REPLACE(
        REGEXP_REPLACE(
          CASE 
            -- Count semicolons to determine if there are more than 5 entries
            WHEN (LENGTH(cd.endpoint_names::text) - LENGTH(REPLACE(cd.endpoint_names::text, ';', ''))) / LENGTH(';') >= 5 THEN
              -- Take portion up to the 5th semicolon and add "[more]"
              SUBSTRING(
                cd.endpoint_names::text, 
                1, 
                COALESCE(NULLIF(STRPOS(
                  SUBSTRING(
                    cd.endpoint_names::text,
                    COALESCE(NULLIF(STRPOS(
                      SUBSTRING(
                        cd.endpoint_names::text,
                        COALESCE(NULLIF(STRPOS(
                          SUBSTRING(
                            cd.endpoint_names::text,
                            COALESCE(NULLIF(STRPOS(cd.endpoint_names::text, ';'), 0), 0) + 1
                          ), 
                          ';'
                        ), 0), 0) + 1
                      ), 
                      ';'
                    ), 0), 0) + 1
                  ), 
                  ';'
                ), 0), LENGTH(cd.endpoint_names::text))
              ) || ' [more]'
            ELSE cd.endpoint_names::text
          END,
          '\\{|\\}', '', 'g'  -- Remove curly braces
        ),
        '"', '', 'g'  -- Remove double quotes
      ),
      '\\\\', '', 'g'  -- Remove escape backslashes
    )
END AS condensed_endpoint_names,
    -- Calculate other metrics
    COUNT(*) OVER (PARTITION BY cd.url) AS num_contacts,
    CASE 
      WHEN cd.contact_name IS NOT NULL OR cd.contact_type IS NOT NULL OR cd.contact_value IS NOT NULL 
      THEN TRUE ELSE FALSE 
    END AS has_contact,
    ROW_NUMBER() OVER (PARTITION BY cd.url ORDER BY cd.contact_preference) AS contact_rank
  FROM contact_data cd
)
SELECT *
FROM endpoints_with_metrics;

-- Create necessary indexes
CREATE UNIQUE INDEX mv_contact_information_uniq
  ON mv_contact_information (url, requested_fhir_version, COALESCE(contact_rank, -1));
CREATE INDEX mv_contact_information_fhir_version_idx 
  ON mv_contact_information (fhir_version);
CREATE INDEX mv_contact_information_vendor_name_idx 
  ON mv_contact_information (vendor_name);
CREATE INDEX mv_contact_information_has_contact_idx 
  ON mv_contact_information (has_contact);