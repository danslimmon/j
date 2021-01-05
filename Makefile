test:
	go test ./...
build: test
	go build -o ./bin/j .
