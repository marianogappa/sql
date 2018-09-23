#!/bin/bash
# Intended to be run by docker-compose
apt-get update && apt-get install -y --no-install-recommends mysql-client && rm -rf /var/lib/apt/lists/* && go test -v .
