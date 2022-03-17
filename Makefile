deps:
	go get -t ./...

build:
	go build -o ./release

test-api:
	cd api && go test

test-keys:
	cd key_storages && go test

start:
	docker-compose up -d

stop:
	docker-compose down

console:
	docker exec -it github-authorized-keys sh

build-image:
	docker build . -t springboard/github-authorized-keys
