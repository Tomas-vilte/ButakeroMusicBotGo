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
	@echo "Ejecutando pruebas en paquetes internos..."
	@go test ./internal/...

	@echo "Ejecutando pruebas en lambdas/discord_notifications..."
	@cd lambdas/discord_notifications && go test ./...

	@echo "Ejecutando pruebas en lambdas/music_download..."
	@cd lambdas/music_download && go test ./...

	@echo "Ejecutando pruebas en lambdas/process_event..."
	@cd lambdas/process_event && go test ./...

	@echo "Ejecutando pruebas en lambdas/lambda_ecs_job_sender..."
	@cd lambdas/lambda_ecs_job_sender && go test ./...

	@echo "Ejecutando pruebas en ecs/process_audio..."
	@cd ecs/process_audio && go test ./...

