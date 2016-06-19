#!/bin/sh

# Fetching and installing all dependencies
# and running server on port 8080. Using this
# script to run the application inside docker
# container.

godep restore

echo 'Running Tests'
go test ./...

cd cmd/
echo 'Running Application'
go run main.go -logtostderr=true -imConvert=/usr/bin/convert