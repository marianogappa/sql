all: build
OS = $(shell uname | tr [:upper:] [:lower:])
ARTIFACT = sql

build: GOOS ?= ${OS}
build: GOARCH ?= amd64
build: clean test
		GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o ${ARTIFACT} -a .

clean: cleanmac
		rm -f ${ARTIFACT}

cleanmac:
		find . -name '*.DS_Store' -type f -delete

test:
		go test

run: build
	./${ARTIFACT}

release-linux: TAG ?= latest
release-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
	tar -cf ${ARTIFACT}-linux.tar ${ARTIFACT}
	gzip ${ARTIFACT}-linux.tar
	rm -rf ${ARTIFACT}

release-darwin: TAG ?= latest
release-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
	tar -cf ${ARTIFACT}-darwin.tar ${ARTIFACT}
	gzip ${ARTIFACT}-darwin.tar
	rm -rf ${ARTIFACT}

release: TAG ?= latest
release: release-linux release-darwin
	git tag ${TAG}
