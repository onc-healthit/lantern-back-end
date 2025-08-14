BEGIN;

-- 1. Database table to store payer registration contact details
CREATE TABLE payers (
    id SERIAL PRIMARY KEY,
    contact_name VARCHAR(500),
    contact_email VARCHAR(500) NOT NULL,
    submission_time TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_email UNIQUE(contact_email)
);

-- 2. Database table to store FHIR endpoints provided by Payer Organizations
CREATE TABLE payer_endpoints (
    id SERIAL PRIMARY KEY,
    payer_id INTEGER REFERENCES payers(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    name VARCHAR(500), -- Organization name of FHIR endpoint
    edi_id INTEGER,    -- Payer ID of the Organization
    address JSONB,     -- Address of the Organization
    is_persisted BOOLEAN DEFAULT FALSE, -- Whether the FHIR endpoint has been added to fhir_endpoints or not
    user_facing_url VARCHAR(500), -- Branding website of the endpoint
    validation_result BOOLEAN, -- Whether the payer registration data validations were successful or not
    validation_comments VARCHAR(500), -- Description of the validation results
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. Database table to store additional payer information required by the tables on Endpoints and Organizations tabs
CREATE TABLE payer_info (
    id SERIAL PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    edi_id INTEGER, -- Payer ID of the Organization
    endpoint_type VARCHAR(500), -- Type of Payer Endpoint. Values will include the four values from the "Type of Endpoint" dropdown in the self-registration form
    address JSONB, -- Address of the Organization
    user_facing_url VARCHAR(500), -- Branding website of the endpoint
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Add triggers for updating timestamps
CREATE TRIGGER set_timestamp_payers
    BEFORE UPDATE ON payers
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_payer_endpoints
    BEFORE UPDATE ON payer_endpoints
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_payer_info
    BEFORE UPDATE ON payer_info
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

-- Add indexes for better performance
CREATE INDEX idx_payers_email ON payers(contact_email);
CREATE INDEX idx_payer_endpoints_payer_id ON payer_endpoints(payer_id);
CREATE INDEX idx_payer_endpoints_url ON payer_endpoints(url);
CREATE INDEX idx_payer_info_url ON payer_info(url);
CREATE INDEX idx_payer_info_edi_id ON payer_info(edi_id);

COMMIT;