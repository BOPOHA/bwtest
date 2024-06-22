
all:
	go build -o bwtest cmd/worker/main.go
run:
	go run cmd/worker/main.go
