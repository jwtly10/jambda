# Specify the default target that will be run when no arguments are given to make.
.PHONY: all
all: test swag run

# Target to generate Swagger documentation using swag init.
.PHONY: swag
swag:
	swag init

# Target to run the Go application.
.PHONY: run
run:
	go run main.go

# Target to run Go tests
.PHONY: test
test: 
	go test ./...

# Target to build the Go application executable.
.PHONY: build
build:
	go build -o myapp main.go

# Target to clean up any binaries or other files.
.PHONY: clean
clean:
	rm -f myapp

