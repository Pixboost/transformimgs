#!/usr/bin/env bash

# To generate schema and api.md:
#
# Install go-swagger (https://goswagger.io) version 0.26.1
# `./generate-api-docs.sh`


set -e

swagger generate spec -w cmd/ -o ./swagger.json
swagger generate markdown --output=api.md