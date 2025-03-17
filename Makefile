# Target para correr todas las pruebas
test: unit-test integration-test

# Target para correr solo las pruebas unitarias
unit-test:
	@echo "Ejecutando pruebas unitarias en microservices/audio_processor"
	@cd microservices/audio_processor && go test -v ./...

	@echo "Ejecutando pruebas unitarias en microservices/butakero_bot"
	@cd microservices/butakero_bot && go test -v ./...

# Target para correr solo las pruebas de integración
integration-test:
	@echo "Ejecutando pruebas de integración en audio_processor"
	@cd microservices/audio_processor && go test -v -tags=integración ./...

	@echo "Ejecutando pruebas de integración en butakero_bot"
	@cd microservices/butakero_bot && go test -v -tags=integración ./...

# Targets predeterminados
.PHONY: test unit-test integration-test
