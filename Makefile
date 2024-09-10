test:
	@echo "Ejecutando pruebas unitarias en internal..."
	@go test ./internal/...

	@echo "Ejecutando pruebas unitarias en microservices/audio_processor/tests/unit..."
	@cd microservices/audio_processor/tests/unit && go test -v ./...

integration-test:
	@echo "Ejecutando pruebas de integración en tests/integration..."
	@cd microservices/audio_processor/tests/integration && go test -v ./...

# Targets predeterminados
.PHONY: test integration-test
