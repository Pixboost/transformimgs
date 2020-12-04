#!/bin/sh

set -e

go mod vendor

echo 'Running Tests'
go test $(go list ./... | grep -v /vendor/) -v -bench . -benchmem
