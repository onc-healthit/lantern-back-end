BEGIN;

DROP TABLE IF EXISTS fhir_endpoint_organization_url;

CREATE TABLE fhir_endpoint_organization_url (
	org_id INT,
	org_url VARCHAR(500)
);

CREATE INDEX idx_fhir_endpoint_organization_url_org_id ON fhir_endpoint_organization_url (org_id);

COMMIT;