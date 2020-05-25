MAKEFLAGS += --warn-undefined-variables

current_dir := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: %

all: clean docker-build build

build:
	docker run -v $(current_dir):/go/src/github.com/hebestreit/clipboard-yt-dl clipboard-yt-dl-build  github.com/hebestreit/clipboard-yt-dl/cmd/clipboard-yt-dl bin

docker-build:
	docker build docker/. -t clipboard-yt-dl-build

clean:
	rm -rf clipboard-yt-dl