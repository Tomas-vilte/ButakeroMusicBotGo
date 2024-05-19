# Variables
BUILD_DIR = build
SRC_DIR = cmd
EXECUTABLE = main
BINARY_NAME = bootstrap
ZIP_NAME := deployment.zip

# Comandos
build:
	@echo "Compilando el código..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)/$(EXECUTABLE).go

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

package: clean
	@echo "Compilando el código..."
	go build -ldflags="-s -w" -o bootstrap main.go 
	@echo "Empaquetando el binario en un archivo ZIP..."
	zip -q $(ZIP_NAME) $(BINARY_NAME)
	@echo "Archivo ZIP generado con éxito: $(ZIP_NAME)"
	rm -rf $(BINARY_NAME)
