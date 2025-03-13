BEGIN;
-- LANTERN-825: Update the history trigger to only insert a new row if the data has changed
-- Drop the existing trigger and function
DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info CASCADE;
DROP FUNCTION IF EXISTS add_fhir_endpoint_info_history() CASCADE;

-- Create new function with metadata awareness and OR logic
CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $fhir_endpoints_info_historys$
BEGIN
    -- For INSERT/DELETE operations, always create history
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'D', now(), user, OLD.*;
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'I', now(), user, NEW.*;
        RETURN NEW;
    END IF;

    -- For UPDATE operations, check if anything significant changed
    IF (
        NEW.id IS DISTINCT FROM OLD.id OR
        NEW.healthit_mapping_id IS DISTINCT FROM OLD.healthit_mapping_id OR
        NEW.vendor_id IS DISTINCT FROM OLD.vendor_id OR
        NEW.url IS DISTINCT FROM OLD.url OR
        NEW.tls_version IS DISTINCT FROM OLD.tls_version OR
        NEW.mime_types IS DISTINCT FROM OLD.mime_types OR
        NEW.capability_statement::text IS DISTINCT FROM OLD.capability_statement::text OR
        NEW.validation_result_id IS DISTINCT FROM OLD.validation_result_id OR
        NEW.included_fields::text IS DISTINCT FROM OLD.included_fields::text OR
        NEW.operation_resource::text IS DISTINCT FROM OLD.operation_resource::text OR
        NEW.supported_profiles::text IS DISTINCT FROM OLD.supported_profiles::text OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        NEW.smart_response::text IS DISTINCT FROM OLD.smart_response::text OR
        NEW.requested_fhir_version IS DISTINCT FROM OLD.requested_fhir_version OR
        NEW.capability_fhir_version IS DISTINCT FROM OLD.capability_fhir_version
    ) THEN
        INSERT INTO fhir_endpoints_info_history 
        SELECT 'U', now(), user, NEW.*;
    END IF;

    RETURN NEW;
END;
$fhir_endpoints_info_historys$ LANGUAGE plpgsql;

CREATE TRIGGER add_fhir_endpoint_info_history_trigger
AFTER INSERT OR UPDATE OR DELETE on fhir_endpoints_info
FOR EACH ROW
WHEN (current_setting('metadata.setting', 't') IS NULL OR current_setting('metadata.setting', 't') = 'FALSE')
EXECUTE PROCEDURE add_fhir_endpoint_info_history();

COMMIT;