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
	exec ./scripts/populatedb.sh


test:
	go test -covermode=atomic -race -count=1 -p 1 ./...

test_int:
	go test -covermode=atomic -race -count=1 -p 1 -tags=integration ./...

test_e2e:
	docker-compose down
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down

test_all:
	make stop
	docker-compose up -d --build
	make test || exit $?
	make test_int || exit $?
	make stop
	make test_e2e || exit $?