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

# Run with CPU profiling
profile:
	go run . --profile

# Run debugger with CPU profiling
profile-dbg:
	go run . --debug --profile

# View profile as flamegraph (run after 'make profile')
flamegraph:
	go tool pprof -http=:8080 cpu.prof

# Generate flamegraph SVG
flamegraph-svg:
	go tool pprof -output=flamegraph.svg -svg cpu.prof