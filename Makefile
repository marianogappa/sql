all: build
OS = $(shell uname | tr [:upper:] [:lower:])
ARTIFACT = sql

build: GOOS ?= ${OS}
build: clean test
		GOOS=${GOOS} GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .

clean: cleanmac
		rm -f ${ARTIFACT}

cleanmac:
		find . -name '*.DS_Store' -type f -delete

test:
	docker-compose --file test-docker-compose.yml up --abort-on-container-exit --force-recreate --renew-anon-volumes

run: build
	./${ARTIFACT}
