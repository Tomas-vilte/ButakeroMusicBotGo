# Variables
BINARY_NAME=bootstrap
AWS_REGION=us-east-1

# Compilar el código Go
build:
	go build -ldflags="-s -w" -o $(BINARY_NAME) cmd/main.go

package:
	zip -r $(BINARY_NAME).zip $(BINARY_NAME)

# Subir el archivo compilado a S3
upload: build
	aws s3 cp $(BINARY_NAME).zip s3://$(S3_BUCKET)/$(S3_FOLDER)/$(BINARY_NAME).zip --region $(AWS_REGION)

# Limpiar los archivos compilados
clean:
	rm -f $(BINARY_NAME)
