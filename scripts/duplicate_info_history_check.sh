#!/bin/sh

docker exec --workdir /go/src/app/cmd/historycleanup lantern-back-end-endpoint_manager-1 go run main.go || echo "Failed to run duplicate info history check script"