# Variables
BUILD_DIR = build
SRC_DIR = cmd
EXECUTABLE = myapp
BINARY_NAME=myapp
ZIP_FILE = $(BUILD_DIR)/lambda.zip

# Comandos
build:
	@echo "Compilando el código..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/main.go

test:
	@echo "Ejecutando pruebas..."
	@go test ./...

run: build
	@echo "Ejecutando la aplicación..."
	@$(BUILD_DIR)/$(EXECUTABLE)

clean:
	@echo "Limpiando el directorio de compilación..."
	rm -rf $(BUILD_DIR)

deps:
	@echo "Instalando dependencias..."
	go mod tidy

package: build
	@echo "Empaquetando la aplicación para AWS Lambda..."
	zip -j $(ZIP_FILE) $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Empaquetado completado: $(ZIP_FILE)"

.PHONY: build test run clean deps package
