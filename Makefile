HASH := $(shell git describe --tags --always)
COMMIT_DATE := $(shell git log -1 --format=%cd --date=short)
VERSION := $(HASH)-$(COMMIT_DATE)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

qwe: *.go
	go build $(LDFLAGS) -v .

clean:
	-rm qwe

install: qwe
	cp qwe ~/.local/bin/
