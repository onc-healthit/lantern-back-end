#!/bin/sh

csv_file="../duplicateInfoHistoryIds.csv"
DB_NAME=lantern
DB_USER=lantern

# Check if the file exists
if [ ! -f "$csv_file" ]; then
    echo "File $csv_file not found!"
    exit 1
fi

# Initial a variable that will hold the validation_result_id from the previous entry.
VAL_RES_ID=-1

while IFS=',' read -r col1 col2 col3 col4; do

    # If the validation_result_id is not 0 and not already processed, then perform the deletion
    if [ "${col4}" -ne "0" ] && [ "${VAL_RES_ID}" -ne "${col4}" ]; then
        
        VAL_RES_ID=$col4
        
        # Check whether there are entries in the history table having the given validation_result_id and operation = 'I'
        QUERY=$(echo "SELECT COUNT(*) FROM fhir_endpoints_info_history WHERE date_trunc('minute', entered_at) <= date_trunc('minute', date('$col2')) AND operation IN ('I', 'U') AND validation_result_id='$col4';")
        COUNT=$(docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error counting entries from the history table"
        
        # Delete corresponding entries from the validations and validation_results tables ONLY IF the count is zero.
        NUMBER=$(echo ${COUNT} | tr -cd '[[:digit:]]')
        if [ "${NUMBER}" -eq "0" ]; then  
            echo "($(date)) Deleting entries from the validations table for validation_result_id: $col4"
            
            # Delete corresponding entry from the validations table
            QUERY=$(echo "DELETE FROM validations WHERE validation_result_id = '$col4';")
            (docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error deleting entry from the validations table"

            # Check whether there are entries in the info table having the given validation_result_id
            QUERY=$(echo "SELECT COUNT(*) FROM fhir_endpoints_info WHERE validation_result_id='$col4';")
            COUNT=$(docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error counting entries from the validation_results table"
            
            # Delete corresponding entry from the validation results table ONLY IF the count is zero.
            NUMBER=$(echo ${COUNT} | tr -cd '[[:digit:]]')
            if [ "${NUMBER}" -eq "0" ]; then  
                echo "($(date)) Deleting entries from the validation_results table for id: $col4"
            
                QUERY=$(echo "DELETE FROM validation_results WHERE id = '$col4';")
                (docker exec -t lantern-back-end_postgres_1 psql -t -U${DB_USER} -d ${DB_NAME} -c "${QUERY}") || echo "Error deleting entry from the validation_results table"    
            fi
        fi
    fi
done < "$csv_file"

echo "Validation data cleanup complete."
