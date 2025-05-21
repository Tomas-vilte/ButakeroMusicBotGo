package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"io"
	"time"
)

type audioDownloaderService struct {
	downloader    ports.Downloader
	encoder       ports.AudioEncoder
	encodeOptions *model.EncodeOptions
	maxAudioSize  int
	log           logger.Logger
}

func NewAudioDownloaderService(d ports.Downloader, e ports.AudioEncoder, l logger.Logger, encodeOptions *model.EncodeOptions) ports.AudioDownloadService {
	return &audioDownloaderService{
		downloader:    d,
		encoder:       e,
		log:           l,
		encodeOptions: encodeOptions,
		maxAudioSize:  100 * 1024 * 1024,
	}
}

func (ad *audioDownloaderService) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	log := ad.log.With(
		zap.String("component", "AudioDownloaderService"),
		zap.String("method", "DownloadAndEncode"),
		zap.String("url", url),
	)
	startTime := time.Now()

	reader, err := ad.downloader.DownloadAudio(ctx, url)
	if err != nil {
		log.Error("Error en descarga", zap.Error(err))
		return nil, err
	}

	session, err := ad.encoder.Encode(ctx, reader, ad.encodeOptions)
	if err != nil {
		log.Error("Error en codificación", zap.Error(err))
		return nil, err
	}
	defer session.Cleanup()

	buffer, err := ad.readAudioFramesToBuffer(session)
	if err != nil {
		log.Error("Error al leer frames", zap.Error(err))
		return nil, err
	}

	log.Info("Audio procesado",
		zap.Int("size_bytes", buffer.Len()),
		zap.Duration("duration", time.Since(startTime)),
	)
	return buffer, nil
}

func (ad *audioDownloaderService) readAudioFramesToBuffer(session ports.EncodeSession) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	for {
		if buffer.Len() > ad.maxAudioSize {
			return nil, fmt.Errorf("el tamaño máximo de audio (%d bytes) ha sido superado", ad.maxAudioSize)
		}
		frame, err := session.ReadFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		_, err = buffer.Write(frame)
		if err != nil {
			return nil, err
		}
	}
	return &buffer, nil
}
