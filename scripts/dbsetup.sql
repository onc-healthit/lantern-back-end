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
    tls_version             VARCHAR(500),
    mime_types              VARCHAR(500)[],
    http_response           INTEGER,
    errors                  VARCHAR(500),
    organization_name       VARCHAR(500),
    fhir_version            VARCHAR(500),
    authorization_standard  VARCHAR(500),
    vendor                  VARCHAR(500),
    location                JSONB, -- location of IP address from ipstack.com.
    capability_statement    JSONB,
    validation              JSONB,
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

CREATE TABLE endpoint_organization (
    endpoint_id INT REFERENCES fhir_enpoints,
    organization_id INT REFERENCES npi_organizations,
    PRIMARY KEY (endpoint_id, organization_id)
)

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
