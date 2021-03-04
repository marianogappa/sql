FROM golang:1.16

RUN apt-get update && apt-get install -y --no-install-recommends default-mysql-client postgresql-client && rm -rf /var/lib/apt/lists/*

ENV GO111MODULE=off

ENTRYPOINT [ "go", "test", "-v", "." ]
