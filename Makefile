ARTIFACT = sql

all: test build

build: export CGO_ENABLED=0
build:
	go build -o ${ARTIFACT} .

test:
	docker-compose run --rm test

clean:
	docker-compose down -v
	docker-compose rm test
