run:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" .

dbg:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" . --debug

build:
	go build -ldflags "-s -w -H=windowsgui -extldflags=-static -extldflags=-lucrt" .