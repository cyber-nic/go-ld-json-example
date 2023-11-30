build:
	go build -o ./example *.go

run:
	go run ./src

test:
	go test -v ./...
