test: unit-test integration-test

unit-test:
	@echo "Ejecutando pruebas unitarias en audio_processor"
	@cd audio_processor && go test -v ./...

	@echo "Ejecutando pruebas unitarias en butakero_bot"
	@cd butakero_bot && go test -v ./...

integration-test:
	@echo "Ejecutando pruebas de integración en audio_processor"
	@cd audio_processor && go test -v -tags=integración ./...

	@echo "Ejecutando pruebas de integración en butakero_bot"
	@cd butakero_bot && go test -v -tags=integración ./...

.PHONY: test unit-test integration-test
