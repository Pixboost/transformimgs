#!/bin/sh

# Fetching and installing all dependencies
# and running server on port 8080. Using this
# script to run the application inside docker
# container.

godep restore

cd cmd/
go run main.go -logtostderr=true -imConvert=/usr/bin/convert