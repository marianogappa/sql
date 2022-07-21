FROM golang:1.11

RUN apt-get update && apt-get install -y --no-install-recommends default-mysql-client postgresql-client && rm -rf /var/lib/apt/lists/*

ENTRYPOINT [ "go", "test", "-v", "." ]
