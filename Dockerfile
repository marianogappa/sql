FROM golang:1.11

RUN apt-get update && apt-get install -y --no-install-recommends mysql-client && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y postgresql-client

ENTRYPOINT [ "go", "test", "-v", "." ]
