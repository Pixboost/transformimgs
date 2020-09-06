#!/bin/sh

set -e

dep ensure

echo 'Running Tests'
go test $(go list ./... | grep -v /vendor/) -v -bench . -benchmem
