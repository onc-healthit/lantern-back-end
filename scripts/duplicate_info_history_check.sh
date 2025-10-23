#!/bin/sh

docker exec --workdir /go/src/app/cmd/historycleanup lantern-back-end_endpoint_manager_1 go run main.go || echo "Failed to run duplicate info history check script"