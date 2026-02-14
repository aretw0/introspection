.PHONY: tidy vet test coverage serve-docs example-basic example-generic

# Ensure dependencies are clean
tidy:
	go mod tidy

# Run vet tool
vet:
	go vet ./...

# Run all tests with race detection
test:
	go test -race -timeout 60s ./...

# Run coverage tests
coverage:
	go test -race -timeout 60s -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run local Go documentation server
serve-docs:
	go tool godoc -http=:6060

# Run basic example
example-basic:
	cd examples/basic && go run main.go

# Run generic example
example-generic:
	cd examples/generic && go run main.go
