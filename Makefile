.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 run --disable errcheck ./...

.PHONY: test
test: fmt lint
	go test -v -race ./...

.PHONY: cover
cover:
	go test -v -race -coverprofile=c.out ./...
	go tool cover -html=c.out -o c.html
	open c.html

.PHONY: up
up:
	docker compose up --build

.PHONY: down
down:
	docker compose down
