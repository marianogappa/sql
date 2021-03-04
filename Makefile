all: build
OS = $(shell uname | tr [:upper:] [:lower:])
ARTIFACT = sql

build: GOOS ?= ${OS}
build: test
		GO111MODULE=off GOOS=${GOOS} GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .

test:
	docker-compose --file test-docker-compose.yml up --abort-on-container-exit --force-recreate --renew-anon-volumes

run: build
	./${ARTIFACT}
