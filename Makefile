# Target para correr todas las pruebas
test: unit-test

# Target para correr solo las pruebas unitarias
unit-test:
	@echo "Ejecutando pruebas unitarias en internal..."
	@go test ./internal/...

	@echo "Ejecutando pruebas unitarias en microservices/audio_processor/tests/unit..."
	@cd microservices/audio_processor && go test -v ./...

# Targets predeterminados
.PHONY: test unit-test integration-test
