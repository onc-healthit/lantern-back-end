#!/bin/sh

set -e

# get endpoint data
cd cmd/endpointpopulator

jq -c '.[]' /etc/lantern/resources/EndpointResourcesList.json | while read endpoint; do
    NAME=$(echo $endpoint | jq -c -r '.EndpointName')
    FORMAT=$(echo $endpoint | jq -c -r '.FormatType')
    FILENAME=$(echo $endpoint | jq -c -r '.FileName')
    LISTURL=$(echo $endpoint | jq -c -r '.URL')

    go run main.go /etc/lantern/resources/$FILENAME $FORMAT $NAME $LISTURL
done

# Only use the line below that populates the database with CareEvolution for development
# go run main.go /etc/lantern/resources/CareEvolutionEndpointSources.json CareEvolution

cd ..

# get CHPL info into db
cd chplquerier
go run main.go
cd ..

# get NPPES contact (endpoint) pfile into db
cd nppescontactpopulator
go run main.go /etc/lantern/resources/endpoint_pfile.csv
cd ..


# get NPPES org pfile data into db
cd nppesorgpopulator
go run main.go /etc/lantern/resources/npidata_pfile.csv
cd ../endpointlinker
go run main.go
cd ..

# run data validation to ensure number of endpoints does not exceed maximum for query interval
cd datavalidation
go run main.go
cd ..