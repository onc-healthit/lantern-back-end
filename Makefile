run:
	docker-compose up --build

run_prod:
	docker-compose -f docker-compose.yml up --build

stop:
	docker-compose down

stop_prod:
	docker-compose -f docker-compose.yml down

clean:
	docker-compose down --rmi all -v
	docker-compose -f docker-compose.yml down --rmi all -v
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml down --rmi all -v
	
test:
	go test -covermode=count -count=1 ./...

test_int:
	go test -covermode=count -count=1 -tags=integration ./...

test_e2e:
	docker-compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.test.yml up --build --abort-on-container-exit
