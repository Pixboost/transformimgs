#!/bin/sh

set -e

go run cmd/main.go -imConvert=/usr/local/bin/convert -imIdentify=/usr/local/bin/identify $@
