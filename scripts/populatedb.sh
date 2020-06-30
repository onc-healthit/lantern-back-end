#!/bin/sh

set -e

# get endpoint data
cd cmd/endpointpopulator
go run main.go /etc/lantern/resources/CernerEndpointSources.json Cerner
go run main.go /etc/lantern/resources/EpicEndpointSources.json Epic
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