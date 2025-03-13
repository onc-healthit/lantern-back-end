BEGIN;
-- LANTERN-825: Update the history trigger to only insert a new row if the data has changed
-- Drop the new trigger and function
DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info CASCADE;
DROP FUNCTION IF EXISTS add_fhir_endpoint_info_history() CASCADE;

-- Restore original function
CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $fhir_endpoints_info_historys$
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
$fhir_endpoints_info_historys$ LANGUAGE plpgsql;

-- Restore original trigger
CREATE TRIGGER add_fhir_endpoint_info_history_trigger
AFTER INSERT OR UPDATE OR DELETE on fhir_endpoints_info
FOR EACH ROW
WHEN (current_setting('metadata.setting', 't') IS NULL OR current_setting('metadata.setting', 't') = 'FALSE')
EXECUTE PROCEDURE add_fhir_endpoint_info_history();

COMMIT;