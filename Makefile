# Variables
BUILD_DIR = build
SRC_DIR = cmd
EXECUTABLE = main
BINARY_NAME = bootstrap

# Comandos
build:
	@echo "Compilando el código..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/$(EXECUTABLE).go

test:
	@echo "Ejecutando pruebas..."
	@go test ./...

