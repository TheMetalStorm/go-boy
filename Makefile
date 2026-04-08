# Run in regular mode
run:
	go run .

# Start with visual debugger
dbg:
	go run . --debug

# Run tests specified in main.go
test:
	go run . --test

# Enable Logging in Gameboy-Doctor format
log:
	go run . --log

# Build in regular mode
build:
	go build .