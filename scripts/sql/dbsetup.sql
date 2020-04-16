CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE fhir_endpoints (
    id                      SERIAL PRIMARY KEY,
    url                     VARCHAR(500) UNIQUE,
    organization_name       VARCHAR(500),
    list_source             VARCHAR(500),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE npi_organizations (
    id               SERIAL PRIMARY KEY,
    npi_id			     VARCHAR(500) UNIQUE,
    name             VARCHAR(500),
    secondary_name   VARCHAR(500),
    location         JSONB,
    taxonomy 		     VARCHAR(500), -- Taxonomy code mapping: http://www.wpc-edi.com/reference/codelists/healthcare/health-care-provider-taxonomy-code-set/
    normalized_name      VARCHAR(500),
    normalized_secondary_name   VARCHAR(500),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE healthit_products (
    id                      SERIAL PRIMARY KEY,
    name                    VARCHAR(500),
    version                 VARCHAR(500),
    developer               VARCHAR(500),
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
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT healthit_product_info UNIQUE(name, version)
);

CREATE TABLE fhir_endpoints_info (
    id                      SERIAL PRIMARY KEY,
    fhir_endpoint_id        INT REFERENCES fhir_endpoints(id) ON DELETE CASCADE,
    healthit_product_id     INT REFERENCES healthit_products(id) ON DELETE SET NULL,
    -- TODO: remove once vendor table available
    vendor                  VARCHAR(500),
    -- TODO: uncomment once vendor table available
    -- vendor_id            INT REFERENCES vendors(id), 
    tls_version             VARCHAR(500),
    mime_types              VARCHAR(500)[],
    http_response           INTEGER,
    errors                  VARCHAR(500),
    capability_statement    JSONB,
    validation              JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE endpoint_organization (
    endpoint_id INT REFERENCES fhir_endpoints (id) ON DELETE CASCADE,
    organization_id INT REFERENCES npi_organizations (id) ON DELETE CASCADE,
    confidence NUMERIC (5, 3),
    CONSTRAINT endpoint_org PRIMARY KEY (endpoint_id, organization_id)
);

CREATE INDEX fhir_endpoint_url_index ON fhir_endpoints (url);

CREATE TRIGGER set_timestamp_fhir_endpoints
BEFORE UPDATE ON fhir_endpoints
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_npi_organization
BEFORE UPDATE ON npi_organizations
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_healthit_products
BEFORE UPDATE ON healthit_products
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE or REPLACE VIEW org_mapping AS
SELECT endpts.url, endpts.vendor, endpts.organization_name AS endpoint_name, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME, orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
LEFT JOIN fhir_endpoints AS endpts ON links.endpoint_id = endpts.id
LEFT JOIN npi_organizations AS orgs ON links.organization_id = orgs.id;

CREATE or REPLACE VIEW endpoint_export AS
SELECT endpts.url, endpts.vendor, endpts.organization_name AS endpoint_name, endpts.tls_version, endpts.mime_types, endpts.http_response, endpts.capability_statement->>'fhirVersion' AS FHIR_VERSION, endpts.capability_statement->>'publisher' AS PUBLISHER, endpts.capability_statement->'software'->'name' AS SOFTWARE_NAME, endpts.capability_statement->'software'->'version' AS SOFTWARE_VERSION, endpts.capability_statement->'software'->'releaseDate' AS SOFTWARE_RELEASEDATE, orgs.name AS ORGANIZATION_NAME, orgs.secondary_name AS ORGANIZATION_SECONDARY_NAME, orgs.taxonomy, orgs.Location->>'state' AS STATE, orgs.Location->>'zipcode' AS ZIPCODE, links.confidence AS MATCH_SCORE
FROM endpoint_organization AS links
RIGHT JOIN fhir_endpoints AS endpts ON links.endpoint_id = endpts.id
LEFT JOIN npi_organizations AS orgs ON links.organization_id = orgs.id;
