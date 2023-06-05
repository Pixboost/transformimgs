#!/bin/sh

set -e

# Using -buildvcs=false because user in the container is root but mounted volume is owned by the user from the host
# See https://github.blog/2022-04-18-highlights-from-git-2-36/#stricter-repository-ownership-checks

echo 'Running Tests'
go test $(go list -buildvcs=false ./... | grep -v /vendor/) -v -bench . -benchmem -race -coverprofile=coverage.txt -covermode=atomic
go test -fuzz=FuzzCalculateTargetSizeForResize -fuzztime 30s ./img/processor/internal/
go test -fuzz=FuzzCalculateTargetSizeForFit -fuzztime 30s ./img/processor/internal/
go test -fuzz=FuzzHttp_LoadImg -fuzztime 30s ./img/loader/
go test -fuzz=FuzzService_ResizeUrl -fuzztime 30s ./img/

