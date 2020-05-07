#!/bin/sh

set -e

# get endpoint data
cd endpointmanager/cmd/endpointpopulator
#go run main.go ../../../networkstatsquerier/resources/EndpointSources.json CareEvolution
#go run main.go ../../../networkstatsquerier/resources/CernerEndpointSources.json Cerner
#go run main.go ../../../networkstatsquerier/resources/EpicEndpointSources.json Epic
cd ../../..

# get CHPL info into db
cd endpointmanager/cmd/chplquerier
go run main.go
cd ../../..

# get NPPES pfile data into db
echo "Do you have an NPPES pfile downloaded (http://download.cms.gov/nppes/NPI_Files.html) and do you want to load it into the database? (y/Y to continue. anything else to stop)"
read cont
if [ "$cont" = "y" ] || [ "$cont" = "Y" ]; then
    cd endpointmanager/cmd/nppesorgpopulator
    echo "Please enter an absolute path for the NPPES pfile CSV file or the path relative to to this location:"
    pwd
    read nppesdata
    echo "Loading data from $nppesdata..."
    go run main.go $nppesdata
    cd ../endpointlinker
    go run main.go
    cd ../../..
else
    echo "No NPPES pfile data will be loaded."
fi

# get NPPES contact (endpoint) data into db
echo "Do you have an NPPES endpoint (contact) file downloaded (http://download.cms.gov/nppes/NPI_Files.html) and do you want to load it into the database? (y/Y to continue. anything else to stop)"
read endpointload
if [ "$endpointload" = "y" ] || [ "$endpointload" = "Y" ]; then
    cd endpointmanager/cmd/nppescontactpopulator
    echo "Please enter an absolute path for the NPPES endpoint (contact) CSV file or the path relative to to this location:"
    pwd
    read nppescontactdata
    echo "Loading data from $nppescontactdata..."
    go run main.go $nppescontactdata
else
    echo "No NPPES endpoint (contact) data will be loaded."
fi
