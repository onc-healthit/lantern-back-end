#!/bin/sh

set -e

# get endpoint data
cd cmd/endpointpopulator

jq -c '.[]' /go/src/app/resources/EndpointResourcesList.json | while read endpoint; do
    NAME=$(echo $endpoint | jq -c -r '.EndpointName')
    FILENAME=$(echo $endpoint | jq -c -r '.FileName')

    go run main.go /etc/lantern/resources/$FILENAME $NAME
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