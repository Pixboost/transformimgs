#!/bin/sh

set -e

go mod vendor

echo 'Running Tests'
go test $(go list ./... | grep -v /vendor/) -v -bench . -benchmem -race -coverprofile=coverage.txt -covermode=atomic
go test -fuzz=FuzzCalculateTargetSizeForResize -fuzztime 30s ./img/processor/internal/
go test -fuzz=FuzzCalculateTargetSizeForFit -fuzztime 30s ./img/processor/internal/
go test -fuzz=FuzzHttp_LoadImg -fuzztime 30s ./img/loader/
go test -fuzz=FuzzService_Transforms -fuzztime 30s ./img/

