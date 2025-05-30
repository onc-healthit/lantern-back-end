#!/bin/sh

log_file="/etc/lantern/populatedb_logs.txt"
current_datetime=$(date +"%Y-%m-%d %H:%M:%S")

set -e

# get endpoint data
cd cmd/endpointpopulator

# Populates the database with State Medicaid endpoints
go run main.go /etc/lantern/resources/MedicaidState_EndpointSources.json Lantern StateMedicaid false StateMedicaid >> $log_file 2>&1

jq -c '.[]' /etc/lantern/resources/MedicareStateEndpointResourcesList.json | while read endpoint; do
    NAME=$(echo $endpoint | jq -c -r '.EndpointName')
    FORMAT=$(echo $endpoint | jq -c -r '.FormatType')
    FILENAME=$(echo $endpoint | jq -c -r '.FileName')
    LISTURL=$(echo $endpoint | jq -c -r '.URL')
    
    if [ -f "/etc/lantern/resources/$FILENAME" ]; then
        go run main.go /etc/lantern/resources/$FILENAME $FORMAT "${NAME}" true $LISTURL >> $log_file 2>&1
    fi
done

jq -c '.[]' /etc/lantern/resources/EndpointResourcesList.json | while read endpoint; do
    NAME=$(echo $endpoint | jq -c -r '.EndpointName')
    FORMAT=$(echo $endpoint | jq -c -r '.FormatType')
    FILENAME=$(echo $endpoint | jq -c -r '.FileName')
    LISTURL=$(echo $endpoint | jq -c -r '.URL')

    go run main.go /etc/lantern/resources/$FILENAME $FORMAT $NAME false $LISTURL >> $log_file 2>&1
done

# Set start time BEFORE processing CHPL endpoints:
POPULATION_START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "$(date '+%Y-%m-%d %H:%M:%S') - CHPL population started at: $POPULATION_START_TIME" >> $log_file

jq -c '.[]' /etc/lantern/resources/CHPLEndpointResourcesList.json | while read endpoint; do
    NAME=$(echo $endpoint | jq -c -r '.EndpointName')
    FORMAT=$(echo $endpoint | jq -c -r '.FormatType')
    FILENAME=$(echo $endpoint | jq -c -r '.FileName')
    LISTURL=$(echo $endpoint | jq -c -r '.URL')
    
    if [ -f "/etc/lantern/resources/$FILENAME" ]; then
        go run main.go /etc/lantern/resources/$FILENAME $FORMAT "${NAME}" true $LISTURL >> $log_file 2>&1
    fi
done

echo "$(date '+%Y-%m-%d %H:%M:%S') - All CHPL sources processed, starting cleanup..." >> $log_file

# Run stale CHPL data cleanup once after all CHPL sources have been processed
cd ../staledatacleaner
if go run main.go "$POPULATION_START_TIME" >> $log_file 2>&1; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') - Stale data cleanup completed successfully" >> $log_file
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') - WARNING: Stale data cleanup failed - check logs" >> $log_file
    # Continue execution - don't fail the entire process
fi
cd ../endpointpopulator

# Only use the line below that populates the database with CareEvolution for development
# go run main.go /etc/lantern/resources/CareEvolutionEndpointSources.json CareEvolution

cd ..

# get CHPL info into db
cd chplquerier
go run main.go >> $log_file 2>&1
cd ..

# get NPPES contact (endpoint) pfile into db
#cd nppescontactpopulator
#go run main.go /etc/lantern/resources/endpoint_pfile.csv
#cd ..


# get NPPES org pfile data into db
#cd nppesorgpopulator
#go run main.go /etc/lantern/resources/npidata_pfile.csv
#cd ../endpointlinker
#go run main.go
#cd ..

# run data validation to ensure number of endpoints does not exceed maximum for query interval
cd datavalidation
go run main.go >> $log_file 2>&1
cd ..