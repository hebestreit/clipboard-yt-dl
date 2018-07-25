MAKEFLAGS += --warn-undefined-variables

.PHONY: %

all: clean deps build

deps:
	dep ensure

build:
	go build -o clipboard-yt-dl -v -i --ldflags=--s cmd/clipboard-yt-dl/main.go

clean:
	rm -rf clipboard-yt-dl vendor/*