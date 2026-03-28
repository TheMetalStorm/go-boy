# Run in regular mode
run:
	go run -tags rgfw .

# Start with visual debugger
dbg:
	go run -tags rgfw . --debug

# Run tests specified in main.go
test:
	go run -tags rgfw . --test

# Enable Logging in Gameboy-Doctor format
log:
	go run -tags rgfw . --log

# Build in regular mode
build:
	go build -tags rgfw .