#!/usr/bin/env sh
set -ex
cd "$(dirname $0)"
go build -v -i
uglifyjs static/embed.js -m > static/embed.min.js
