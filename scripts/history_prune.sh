#!/bin/sh

SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

docker exec --workdir /go/src/app/cmd/historypruning lantern-back-end_endpoint_manager_1 go run main.go true

