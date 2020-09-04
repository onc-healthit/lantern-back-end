BEGIN;

CREATE TABLE IF NOT EXISTS fhir_endpoints_availability (
    url             VARCHAR(500),
    http_200_count       BIGINT,
    http_all_count       BIGINT,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE fhir_endpoints_info ADD COLUMN availability DECIMAL(64,4);
ALTER TABLE fhir_endpoints_info_history ADD COLUMN availability DECIMAL(64,4);

CREATE TRIGGER set_timestamp_fhir_endpoint_availability
BEFORE UPDATE ON fhir_endpoints_availability
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE OR REPLACE FUNCTION populate_existing_tables_availability_info() RETURNS VOID as $$
    DECLARE 
        i             record;
        okay_count      bigint;
        all_count       bigint;
    BEGIN
        FOR i IN SELECT DISTINCT fhir_endpoints_info_history.url FROM fhir_endpoints_info_history
        LOOP
            SELECT COUNT(*) INTO all_count FROM fhir_endpoints_info_history WHERE url = i.url AND (operation = 'I' OR operation = 'U');
            SELECT COUNT(*) INTO okay_count FROM fhir_endpoints_info_history WHERE url = i.url AND http_response = 200 AND (operation = 'I' OR operation = 'U');
            INSERT INTO fhir_endpoints_availability VALUES (i.url, okay_count, all_count);
            UPDATE fhir_endpoints_info SET availability = (okay_count * 1.0) / all_count WHERE url = i.url;
        END LOOP;
    END
$$ LANGUAGE plpgsql;

SELECT populate_existing_tables_availability_info();

CREATE OR REPLACE FUNCTION update_fhir_endpoint_availability_info() RETURNS TRIGGER AS $fhir_endpoints_availability$
    DECLARE
        okay_count       bigint;
        all_count        bigint;
    BEGIN
        --
        -- Create or update a row in fhir_endpoint_availabilty with new total http and 200 http count 
        -- when an endpoint is inserted or updated in fhir_endpoint_info. Also calculate new 
        -- endpoint availability precentage
        SELECT http_200_count, http_all_count INTO okay_count, all_count FROM fhir_endpoints_availability WHERE url = NEW.url;
        IF  NOT FOUND THEN
            IF NEW.http_response = 200 THEN
                INSERT INTO fhir_endpoints_availability VALUES (NEW.url, 1, 1);
                NEW.availability = 1.00;
                RETURN NEW;
            ELSE
                INSERT INTO fhir_endpoints_availability VALUES (NEW.url, 0, 1);
                NEW.availability = 0.00;
                RETURN NEW;
            END IF;
        ELSE
            IF NEW.http_response = 200 THEN
                UPDATE fhir_endpoints_availability SET http_200_count = okay_count + 1.0, http_all_count = all_count + 1.0 WHERE url = NEW.url;
                NEW.availability := (okay_count + 1.0) / (all_count + 1.0);
                RETURN NEW;
            ELSE
                UPDATE fhir_endpoints_availability SET http_all_count = all_count + 1.0 WHERE url = NEW.url;
                NEW.availability := (okay_count) / (all_count + 1.0);
                RETURN NEW;
            END IF;
        END IF;
    END;
$fhir_endpoints_availability$ LANGUAGE plpgsql;

CREATE TRIGGER update_fhir_endpoint_availability_trigger
BEFORE INSERT OR UPDATE on fhir_endpoints_info
FOR EACH ROW
EXECUTE PROCEDURE update_fhir_endpoint_availability_info();


COMMIT;
