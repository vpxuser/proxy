.PHONY: test build vet tidy clean lint

test:
	go test ./... -v -count=1

build:
	go build ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	go clean
	rm -f proxy proxy.exe

lint:
	golangci-lint run ./...

build-example:
	cd examples/printer && go build -o proxy .
