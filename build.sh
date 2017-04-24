#!/usr/bin/env sh
set -ex
cd "$(dirname $0)"
go build -v -i -o app
strip app -o app.stripped
objcopy --only-keep-debug app app.debug
objcopy --add-gnu-debuglink=app.debug app.stripped
uglifyjs static/embed.js -m > static/embed.min.js
