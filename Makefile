.PHONY: proto run build test docker-up docker-down clean swagger

# Generate protobuf code (requires protoc + plugins installed)
proto:
	protoc \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=proto/gen --grpc-gateway_opt=paths=source_relative \
		-I proto \
		-I third_party/googleapis \
		proto/user.proto

# Generate Swagger docs (requires: go install github.com/swaggo/swag/cmd/swag@latest)
swagger:
	swag init -g cmd/server/main.go -o docs --parseInternal

# Run locally
run:
	go run ./cmd/server

# Build binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# Run tests
test:
	go test -v -race ./...

# Docker Compose
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

# Tidy & clean
tidy:
	go mod tidy

clean:
	rm -rf bin/
