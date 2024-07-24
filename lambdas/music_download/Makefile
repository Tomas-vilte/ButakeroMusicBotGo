.PHONY: build clean deploy

BINARY_NAME=bootstrap
SRC_DIR=cmd
EXECUTABLE=main
S3_BUCKET_NAME=butakero

build:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME) $(SRC_DIR)/$(EXECUTABLE).go

clean:
	rm -f $(BINARY_NAME)

zip: build
	zip $(BINARY_NAME).zip $(BINARY_NAME)

deploy: zip
	aws s3 cp $(BINARY_NAME).zip s3://$(S3_BUCKET_NAME)/$(BINARY_NAME).zip
	aws lambda update-function-code --function-name GoFunctionAudio --s3-bucket $(S3_BUCKET_NAME) --s3-key $(BINARY_NAME).zip
