#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

cd ../endpointmanager/cmd/historypruning
go run main.go true

cd ../../..

docker exec --workdir /go/src/app/cmd/jsonexport lantern-back-end_endpoint_manager_1 go run main.go '/etc/lantern/exportfolder/fhir_endpoints_fields.json'
