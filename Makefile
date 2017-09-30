GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOGET=$(GOCMD) get
BINARY_NAME=mxdisk

all: test build fmt
fmt: 
		$(GOFMT) ./...
build: 
		$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/mxdisk
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
run:
		$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/mxdisk
		./$(BINARY_NAME)
