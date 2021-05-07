BEGIN;

CREATE TABLE validation_results (
    id                      SERIAL PRIMARY KEY
);

ALTER TABLE fhir_endpoints_info 
ADD COLUMN validation_result_id INT REFERENCES validation_results(id) ON DELETE SET NULL;
ALTER TABLE fhir_endpoints_info_history
ADD COLUMN validation_result_id INT REFERENCES validation_results(id) ON DELETE SET NULL;

ALTER TABLE fhir_endpoints_info DROP COLUMN IF EXISTS validation CASCADE;
ALTER TABLE fhir_endpoints_info_history DROP COLUMN IF EXISTS validation CASCADE;

CREATE TABLE IF NOT EXISTS validations (
    rule_name               VARCHAR(500),
    valid                   BOOLEAN,
    expected                VARCHAR(500),
    actual                  VARCHAR(500),
    comment                 VARCHAR(500),
    reference               VARCHAR(500),
    implementation_guide    VARCHAR(500),
    validation_result_id    INT REFERENCES validation_results(id) ON DELETE SET NULL
);

COMMIT;
