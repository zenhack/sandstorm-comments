
all: app.stripped app.debug static/embed.min.js
clean:
	rm -f app app.stripped app.debug static/*.min.js
dev: all
	spk dev -p .sandstorm/sandstorm-pkgdef.capnp:pkgdef

app: $(wildcard *.go)
	go build -v -i -o app
app.stripped: app app.debug
	strip app -o app.stripped
	objcopy --add-gnu-debuglink=app.debug app.stripped
app.debug: app
	objcopy --only-keep-debug app app.debug
static/%.min.js: static/%.js
	uglifyjs $< -m > $@

.PHONY: all clean
