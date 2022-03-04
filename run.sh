#!/bin/sh

set -e

# Fetching and installing all dependencies
# and running server on port 8080. Using this
# script to run the application inside docker
# container.

go mod vendor

cd cmd/
echo 'Running Application'
go run main.go -imConvert=/usr/local/bin/convert -imIdentify=/usr/local/bin/identify $@