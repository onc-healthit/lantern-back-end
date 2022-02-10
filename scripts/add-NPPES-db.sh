#!/bin/sh

set -e

echo "Getting NPPES contact (endpoint) pfile into db"
cd cmd/nppescontactpopulator
go run main.go /etc/lantern/resources/endpoint_pfile.csv
cd ..

echo "Getting NPPES org pfile data into db"
cd nppesorgpopulator
go run main.go /etc/lantern/resources/npidata_pfile.csv
