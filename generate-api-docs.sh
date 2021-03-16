#!/usr/bin/env bash

set -e

swagger generate spec -w cmd/ -o ../swagger.json
swagger generate markdown --output=api.md