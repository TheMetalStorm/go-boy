run:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" .

dbg:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --debug

test:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --test

build:
	go build -ldflags "-s -w -H=windowsgui -extldflags=-static -extldflags=-lucrt" .