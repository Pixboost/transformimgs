#!/bin/sh

# Fetching and installing all dependencies
# and running server on port 8080. Using this
# script to run the application inside docker
# container.

dep ensure

echo 'Running Tests'
go test ./... -imConvert=/usr/bin/convert -imIdentify=/usr/bin/identify

cd cmd/
echo 'Running Application'
go run main.go -imConvert=/usr/bin/convert -imIdentify=/usr/bin/identify $@