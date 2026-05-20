BEGIN;
-- Fix positional column mismatch in add_fhir_endpoint_info_history.
-- The previous implementation used SELECT ... NEW.* which maps columns by position.
-- When columns were added to fhir_endpoints_info via migrations (appended to the end),
-- the positional order diverged from fhir_endpoints_info_history, causing type errors such as:
--   column "mime_types" is of type character varying[] but expression is of type character varying
-- Using explicit column names eliminates the positional dependency.

CREATE OR REPLACE FUNCTION add_fhir_endpoint_info_history() RETURNS TRIGGER AS $fhir_endpoints_info_historys$
BEGIN
    -- For INSERT/DELETE operations, always create history
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO fhir_endpoints_info_history (
            operation, entered_at, user_id,
            id, healthit_mapping_id, vendor_id, url, tls_version, mime_types,
            capability_statement, validation_result_id, included_fields,
            operation_resource, supported_profiles, created_at, updated_at,
            smart_response, metadata_id, requested_fhir_version, capability_fhir_version)
        VALUES (
            'D', now(), user,
            OLD.id, OLD.healthit_mapping_id, OLD.vendor_id, OLD.url, OLD.tls_version,
            OLD.mime_types, OLD.capability_statement, OLD.validation_result_id,
            OLD.included_fields, OLD.operation_resource, OLD.supported_profiles,
            OLD.created_at, OLD.updated_at, OLD.smart_response, OLD.metadata_id,
            OLD.requested_fhir_version, OLD.capability_fhir_version);
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO fhir_endpoints_info_history (
            operation, entered_at, user_id,
            id, healthit_mapping_id, vendor_id, url, tls_version, mime_types,
            capability_statement, validation_result_id, included_fields,
            operation_resource, supported_profiles, created_at, updated_at,
            smart_response, metadata_id, requested_fhir_version, capability_fhir_version)
        VALUES (
            'I', now(), user,
            NEW.id, NEW.healthit_mapping_id, NEW.vendor_id, NEW.url, NEW.tls_version,
            NEW.mime_types, NEW.capability_statement, NEW.validation_result_id,
            NEW.included_fields, NEW.operation_resource, NEW.supported_profiles,
            NEW.created_at, NEW.updated_at, NEW.smart_response, NEW.metadata_id,
            NEW.requested_fhir_version, NEW.capability_fhir_version);
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
        INSERT INTO fhir_endpoints_info_history (
            operation, entered_at, user_id,
            id, healthit_mapping_id, vendor_id, url, tls_version, mime_types,
            capability_statement, validation_result_id, included_fields,
            operation_resource, supported_profiles, created_at, updated_at,
            smart_response, metadata_id, requested_fhir_version, capability_fhir_version)
        VALUES (
            'U', now(), user,
            NEW.id, NEW.healthit_mapping_id, NEW.vendor_id, NEW.url, NEW.tls_version,
            NEW.mime_types, NEW.capability_statement, NEW.validation_result_id,
            NEW.included_fields, NEW.operation_resource, NEW.supported_profiles,
            NEW.created_at, NEW.updated_at, NEW.smart_response, NEW.metadata_id,
            NEW.requested_fhir_version, NEW.capability_fhir_version);
    END IF;

    RETURN NEW;
END;
$fhir_endpoints_info_historys$ LANGUAGE plpgsql;

COMMIT;
