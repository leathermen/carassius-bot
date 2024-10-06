# Define binary names for both projects
CONSUMER_BINARY=consumer-app
PRODUCER_BINARY=producer-app

# Define the main Go files for both projects
CONSUMER_MAIN=./consumer/cmd/main.go
PRODUCER_MAIN=./producer/cmd/main.go

# Build the consumer binary
.PHONY: build-consumer
build-consumer:
	@echo "Building the consumer binary..."
	go build -o $(CONSUMER_BINARY) $(CONSUMER_MAIN)

# Build the producer binary
.PHONY: build-producer
build-producer:
	@echo "Building the producer binary..."
	go build -o $(PRODUCER_BINARY) $(PRODUCER_MAIN)

# Run golangci-lint with no cache for both projects
.PHONY: lint
lint:
	@echo "Running golangci-lint on consumer..."
	(cd ./consumer && golangci-lint cache clean && golangci-lint run)
	@echo "Running golangci-lint on producer..."
	(cd ./producer && golangci-lint cache clean && golangci-lint run)
	@echo "Running golangci-lint on shared..."
	(cd ./shared && golangci-lint cache clean && golangci-lint run)

.PHONY: test
test:
	@echo "Testing consumer..."
	(go test ./consumer/...)
	@echo "Testing producer..."
	(go test ./producer/...)

# Clean up binaries
.PHONY: clean
clean:
	@echo "Cleaning up binaries..."
	rm -f $(CONSUMER_BINARY) $(PRODUCER_BINARY)

# Run the consumer binary
.PHONY: run-consumer
run-consumer: build-consumer
	./$(CONSUMER_BINARY)

# Run the producer binary
.PHONY: run-producer
run-producer: build-producer
	./$(PRODUCER_BINARY)
