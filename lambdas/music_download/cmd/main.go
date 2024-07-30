package main

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/api/youtube_api"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/downloader"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/handler"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/service/provider/youtube_provider"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/service/sqs"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/uploader"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.uber.org/zap"
	"os"
)

func main() {
	cfg := &config.Config{
		BucketName:    os.Getenv("BUCKET_NAME"),
		AccessKey:     os.Getenv("ACCESS_KEY"),
		SecretKey:     os.Getenv("SECRET_KEY"),
		Region:        os.Getenv("REGION"),
		YouTubeApiKey: os.Getenv("YOUTUBE_API_KEY"),
		QueueURL:      os.Getenv("QUEUE_URL"),
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
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
	}))
	download := downloader.NewDownloader(uploaderS3, logger, commandExecutor, "/opt/lambda-layer/bin/yt-dlp")

	youtubeClient, err := youtube_provider.NewRealYouTubeClient(cfg.YouTubeApiKey)
	if err != nil {
		logger.Error("Error al conectarse al cliente de youtube", zap.Error(err))
		panic("Error al conectarse al cliente de youtube")
	}
	youtubeService := youtube_provider.NewYouTubeProvider(logger, youtubeClient)
	youtubeFetcher := youtube_api.NewYoutubeFetcher(logger, youtubeService)
	sqsClient := sqs.NewSQSClient(sess, cfg.QueueURL, logger)
	handlerLambda := handler.NewHandler(download, uploaderS3, logger, youtubeFetcher, sqsClient)

	lambda.Start(func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handlerLambda.HandleEvent(ctx, event)
	})
}
