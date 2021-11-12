BEGIN;

ALTER TABLE fhir_endpoints_metadata ADD COLUMN requested_fhir_version VARCHAR(500) DEFAULT 'None';

ALTER TABLE fhir_endpoints_availability ADD COLUMN requested_fhir_version VARCHAR(500) DEFAULT 'None';

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
                INSERT INTO fhir_endpoints_availability VALUES (NEW.url, 1, 1, NEW.requested_fhir_version);
                NEW.availability = 1.00;
                RETURN NEW;
            ELSE
                INSERT INTO fhir_endpoints_availability VALUES (NEW.url, 0, 1, NEW.requested_fhir_version);
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

COMMIT;