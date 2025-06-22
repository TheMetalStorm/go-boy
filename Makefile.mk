# Run in regular mode
run:
	go run -tags rgfw -ldflags "-extldflags=-static" .

# Start with visual debugger
dbg:
	go run -tags rgfw -ldflags "-extldflags=-static" . --debug

# Run tests specified in main.go
test:
	go run -tags rgfw -ldflags "-extldflags=-static" . --test

# Enable Logging in Gameboy-Doctor format
log:
	go run -tags rgfw -ldflags "-extldflags=-static" . --log

# Build in regular mode
build:
	go build -tags rgfw -ldflags "-s -w -H=windowsgui -extldflags=-static" .