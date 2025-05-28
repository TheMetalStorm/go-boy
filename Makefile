run:
	go run -ldflags "-extldflags=-static -extldflags=-lucrt" .

build:
	go build -ldflags "-s -w -H=windowsgui -extldflags=-static -extldflags=-lucrt" .