.PHONY: start format

start:
	go run *.go

format:
	gofmt -w *.go
	goimports -w *.go
