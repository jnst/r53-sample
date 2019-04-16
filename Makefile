.PHONY: start format

start:
	go run main.go

format:
	gofmt -w *.go
	goimports -w *.go
