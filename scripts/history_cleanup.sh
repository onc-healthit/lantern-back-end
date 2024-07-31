#!/bin/sh

csv_file="../duplicateInfoHistoryIds.csv"
DB_NAME=lantern
DB_USER=lantern

# Check if the file exists
if [ ! -f "$csv_file" ]; then
    echo "File $csv_file not found!"
    exit 1
fi

while IFS=',' read -r col1 col2 col3 col4; do
    DATE=$(date)
    echo "($DATE) Deleting entries for data: $col1, $col2, $col3, $col4"
    
    # Delete entry from the info history table
    QUERY=$(echo "DELETE FROM fhir_endpoints_info_history WHERE url='$col1' AND operation='U' AND requested_fhir_version='$col3' AND entered_at = '$col2';")
    (docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error deleting entry from the info history table"

    # Delete corresponding entry from the validations table
    QUERY=$(echo "DELETE FROM validations WHERE validation_result_id = '$col4';")
    (docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error deleting entry from the validations table"

    # Check whether there are entries in the info table having the given validation_result_id
    QUERY=$(echo "SELECT COUNT(*) FROM fhir_endpoints_info WHERE validation_result_id='$col4';")
    COUNT=$(docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error counting entries from the validation_results table"

    # Delete corresponding entry from the validation results table ONLY IF the count is zero.
    NUMBER=$(echo ${COUNT} | tr -cd '[[:digit:]]')
    if [ "${NUMBER}" -eq "0" ]; then  
        QUERY=$(echo "DELETE FROM validation_results WHERE id = '$col4';")
        (docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error deleting entry from the validation_results table"    
    fi
done < "$csv_file"

echo "Duplicate info history data cleanup complete."
