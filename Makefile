run:
	go run cmd/app/main.go -config config.toml -debug

test:
	go test -v ./...