#!/bin/sh

DB_NAME=lantern
DB_USER=lantern
DB_CONTAINER=lantern-back-end-postgres-1

DATE=$(date)
echo "($DATE) Starting cleanup of orphaned validation_results entries..."

QUERY="
DELETE FROM validation_results vr
WHERE vr.id IS NOT NULL
AND NOT EXISTS (
    SELECT 1 FROM fhir_endpoints_info fei WHERE fei.validation_result_id = vr.id
)
AND NOT EXISTS (
    SELECT 1 FROM fhir_endpoints_info_history feih WHERE feih.validation_result_id = vr.id
);
"

docker exec -t "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "$QUERY"

if [ $? -eq 0 ]; then
  echo "($DATE) Validations Cleanup successful."
else
  echo "($DATE) Error occurred during Validations Cleanup."
fi
