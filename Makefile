run:
	docker-compose up --build

run_prod:
	docker-compose -f docker-compose.yml up --build

stop:
	docker-compose down

stop_prod:
	docker-compose -f docker-compose.yml down

clean:
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Are you sure you want to clean all files? This REMOVES ALL VOLUMES. Type y/Y to continue to clean: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)

	docker-compose down --rmi local -v
	docker-compose -f docker-compose.yml down --rmi local -v
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down --rmi local -v

clean_remote:
	@while [ -z "$$CONTINUE" ]; do \
        read -r -p "Are you sure you want to clean all files? This REMOVES ALL VOLUMES. Type y/Y to continue to clean: " CONTINUE; \
    done ; \
    [ $$CONTINUE = "y" ] || [ $$CONTINUE = "Y" ] || (echo "Exiting."; exit 1;)

	docker-compose down --rmi all -v
	docker-compose -f docker-compose.yml down --rmi all -v
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down --rmi all -v

populatedb:
	exec docker exec -it lantern-back-end_endpoint_manager_1 /etc/lantern/populatedb.sh

lint:
	cd ./capabilityquerier; golangci-lint run -E gofmt
	cd ./networkstatsquerier; golangci-lint run -E gofmt
	cd ./lanternmq; golangci-lint run -E gofmt
	cd ./fhir; golangci-lint run -E gofmt
	cd ./endpointmanager; golangci-lint run -E gofmt
	cd ./capabilityreceiver; golangci-lint run -E gofmt

csv_export:
	cd endpointmanager/cmd/endpointexporter; go run main.go; docker cp lantern-back-end_pg_prometheus_1:/tmp/export.csv ../../../lantern_export_`date +%F`.csv

test:
	cd ./capabilityquerier; go test -covermode=atomic -race -count=1 -p 1 ./...
	cd ./networkstatsquerier; go test -covermode=atomic -race -count=1 -p 1 ./...
	cd ./lanternmq; go test -covermode=atomic -race -count=1 -p 1 ./...
	cd ./fhir; go test -covermode=atomic -race -count=1 -p 1 ./...
	cd ./endpointmanager; go test -covermode=atomic -race -count=1 -p 1 ./...
	cd ./capabilityreceiver; go test -covermode=atomic -race -count=1 -p 1 ./...

test_int:
	cd ./capabilityquerier; go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...
	cd ./networkstatsquerier; go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...
	cd ./lanternmq;	go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...
	cd ./fhir; go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...
	cd ./endpointmanager; go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...
	cd ./capabilityreceiver; go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...

test_e2e:
	docker-compose down
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml up  --scale shinydashboard=0 --build --abort-on-container-exit
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down

test_all:
	make stop
	docker-compose up -d --build
	make test || exit $?
	make test_int || exit $?
	make stop
	make test_e2e || exit $?

update_mods:
	[  -z "$(branch)" ] && echo "No branch name specified, will update gomods to master" || echo "Updating gomods to point to branch $(branch)"
	cd ./e2e; go get github.com/onc-healthit/lantern-back-end/endpointmanager@$(branch); go get github.com/onc-healthit/lantern-back-end/networkstatsquerier@$(branch); go get github.com/onc-healthit/lantern-back-end/capabilityquerier@$(branch); go get github.com/onc-healthit/lantern-back-end/lanternmq@$(branch); go get github.com/onc-healthit/lantern-back-end/capabilityreceiver@$(branch); go mod tidy;
	cd ./networkstatsquerier; go get github.com/onc-healthit/lantern-back-end/lanternmq@$(branch); go get github.com/onc-healthit/lantern-back-end/endpointmanager@$(branch); go mod tidy;
	cd ./capabilityquerier; go get github.com/onc-healthit/lantern-back-end/endpointmanager@$(branch); go get github.com/onc-healthit/lantern-back-end/lanternmq@$(branch); go get github.com/onc-healthit/lantern-back-end/networkstatsquerier@$(branch); go mod tidy;
	cd ./endpointmanager; go get github.com/onc-healthit/lantern-back-end/lanternmq@$(branch); go get github.com/onc-healthit/lantern-back-end/networkstatsquerier@$(branch); go mod tidy;
	cd ./capabilityreceiver; go get github.com/onc-healthit/lantern-back-end/endpointmanager@$(branch); go get github.com/onc-healthit/lantern-back-end/lanternmq@$(branch); go mod tidy;
