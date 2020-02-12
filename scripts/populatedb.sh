#!/bin/sh

set -e

# get endpoint data
cd endpointmanager/cmd/endpointpopulator
go run main.go ../../../networkstatsquerier/resources/EndpointSources.json
cd ../../..

# get CHPL info into db
cd endpointmanager/cmd
go run main.go
cd ../..

# get NPPES data into db
echo "Do you have NPPES data downloaded (http://download.cms.gov/nppes/NPI_Files.html) and do you want to load it into the database? (y/Y to continue. anything else to stop)"
read cont
if [ "$cont" = "y" ] || [ "$cont" = "Y" ]; then
    cd endpointmanager/cmd/nppespopulator
    echo "Please enter an absolute path for the NPPES data CSV file or the path relative to to this location:"
    pwd
    read nppesdata
    echo "Loading data from $nppesdata..."
    go run main.go $nppesdata
else
    echo "No NPPES data will be loaded."
fi