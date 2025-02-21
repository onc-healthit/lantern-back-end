-- LANTERN-825: Update the history trigger to only insert a new row if the data has changed
BEGIN;

-- Drop the existing trigger and function
DROP TRIGGER IF EXISTS add_fhir_endpoint_info_history_trigger ON fhir_endpoints_info CASCADE;
DROP FUNCTION IF EXISTS add_fhir_endpoint_info_history() CASCADE;

-- Create new function with metadata awareness
CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE' AND NEW.id IS NOT DISTINCT FROM OLD.id AND 
        NEW.healthit_mapping_id IS NOT DISTINCT FROM OLD.healthit_mapping_id AND 
        NEW.vendor_id IS NOT DISTINCT FROM OLD.vendor_id AND 
        NEW.url IS NOT DISTINCT FROM OLD.url AND 
        NEW.tls_version IS NOT DISTINCT FROM OLD.tls_version AND 
        NEW.mime_types IS NOT DISTINCT FROM OLD.mime_types AND 
        NEW.capability_statement::text IS NOT DISTINCT FROM OLD.capability_statement::text AND 
        NEW.validation_result_id IS NOT DISTINCT FROM OLD.validation_result_id AND 
        NEW.included_fields::text IS NOT DISTINCT FROM OLD.included_fields::text AND 
        NEW.operation_resource::text IS NOT DISTINCT FROM OLD.operation_resource::text AND 
        NEW.supported_profiles::text IS NOT DISTINCT FROM OLD.supported_profiles::text AND 
        NEW.created_at IS NOT DISTINCT FROM OLD.created_at AND 
        NEW.updated_at IS NOT DISTINCT FROM OLD.updated_at AND 
        NEW.smart_response::text IS NOT DISTINCT FROM OLD.smart_response::text AND 
        NEW.requested_fhir_version IS NOT DISTINCT FROM OLD.requested_fhir_version AND 
        NEW.capability_fhir_version IS NOT DISTINCT FROM OLD.capability_fhir_version AND
        (NEW.metadata_id IS NOT DISTINCT FROM OLD.metadata_id OR 
         (NEW.metadata_id IS DISTINCT FROM OLD.metadata_id AND 
          current_setting('metadata.setting', 't') = 'TRUE'))) THEN
        RETURN NEW;
    END IF;

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
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create new trigger
CREATE TRIGGER add_fhir_endpoint_info_history_trigger 
AFTER INSERT OR UPDATE OR DELETE ON fhir_endpoints_info 
FOR EACH ROW EXECUTE FUNCTION add_fhir_endpoint_info_history();

COMMIT;