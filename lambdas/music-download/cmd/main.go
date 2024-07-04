package main

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/config"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/downloader"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/handler"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/uploader"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
)

func main() {
	cfg := &config.Config{
		BucketName: os.Getenv("BUCKET_NAME"),
		AccessKey:  os.Getenv("ACCESS_KEY"),
		SecretKey:  os.Getenv("SECRET_KEY"),
		Region:     os.Getenv("REGION"),
	}
	logger, err := logging.NewZapLogger(false)
	if err != nil {
		panic("Error creando el logger: " + err.Error())
	}
	uploaderS3, err := uploader.NewS3Uploader(logger, *cfg)
	if err != nil {
		panic("Error creando el uploader: " + err.Error())
	}
	commandExecutor := downloader.NewCommandExecutor()
	download := downloader.NewDownloader(uploaderS3, logger, commandExecutor, "/opt/lambda-layer/bin/yt-dlp")

	handlerLambda := handler.NewHandler(download, uploaderS3, logger)

	lambda.Start(func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handlerLambda.HandleEvent(ctx, event)
	})
}
