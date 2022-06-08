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
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT healthit_product_info UNIQUE(name, version)
);

CREATE TABLE healthit_products_map (
    id SERIAL,
    healthit_product_id INT REFERENCES healthit_products(id) ON DELETE SET NULL
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
    organization_names      VARCHAR(500)[],
    npi_ids                 VARCHAR(500)[],
    list_source             VARCHAR(500),
    versions_response       JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fhir_endpoints_unique UNIQUE(url, list_source)
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
    healthit_mapping_id     INT -- should link to healthit_products_map(id). not using 'reference' because the referenced id might have multiple entries and thus is not a primary key,
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


CREATE TRIGGER set_timestamp_fhir_endpoints
BEFORE UPDATE ON fhir_endpoints
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

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.list_source, endpts.organization_names AS endpoint_names,
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
    endpts_info.requested_fhir_version,
    orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME,
    orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, orgs.npi_id as NPI_ID,
    links.confidence AS MATCH_SCORE, endpts_metadata.availability
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.url = endpts.url
LEFT JOIN fhir_endpoints_info AS endpts_info ON endpts.url = endpts_info.url
LEFT JOIN fhir_endpoints_metadata AS endpts_metadata ON endpts_info.metadata_id = endpts_metadata.id
LEFT JOIN vendors ON endpts_info.vendor_id = vendors.id
LEFT JOIN npi_organizations AS orgs ON links.organization_npi_id = orgs.npi_id;

CREATE INDEX fhir_endpoints_url_idx ON fhir_endpoints (url);
CREATE INDEX fhir_endpoints_info_url_idx ON fhir_endpoints_info (url);
CREATE INDEX fhir_endpoints_info_history_url_idx ON fhir_endpoints_info_history (url);
CREATE INDEX endpoint_organization_url_idx ON endpoint_organization (url);

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