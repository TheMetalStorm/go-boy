run:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" .

dbg:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --debug

test:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --test

log:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --log

build:
	go build -ldflags "-s -w -H=windowsgui -extldflags=-static -extldflags=-lucrt" .