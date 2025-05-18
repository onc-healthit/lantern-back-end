BEGIN;

DELETE FROM validation_results vr
WHERE vr.id IS NOT NULL
AND NOT EXISTS (
    SELECT 1 FROM fhir_endpoints_info fei WHERE fei.validation_result_id = vr.id
)
AND NOT EXISTS (
    SELECT 1 FROM fhir_endpoints_info_history feih WHERE feih.validation_result_id = vr.id
);

COMMIT;