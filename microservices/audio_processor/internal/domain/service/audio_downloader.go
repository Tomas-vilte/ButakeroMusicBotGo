package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"io"
)

type AudioDownloader struct {
	downloader ports.Downloader
	encoder    ports.AudioEncoder
	log        logger.Logger
}

func NewAudioDownloader(d ports.Downloader, e ports.AudioEncoder, l logger.Logger) *AudioDownloader {
	return &AudioDownloader{
		downloader: d,
		encoder:    e,
		log:        l,
	}
}

func (ad *AudioDownloader) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	reader, err := ad.downloader.DownloadAudio(ctx, url)
	if err != nil {
		return nil, errors.ErrDownloadFailed.Wrap(err)
	}

	session, err := ad.encoder.Encode(ctx, reader, encoder.StdEncodeOptions)
	if err != nil {
		return nil, errors.ErrEncodingFailed.Wrap(err)
	}
	defer session.Cleanup()

	return ad.readAudioFramesToBuffer(session)
}

// readAudioFramesToBuffer lee los frames de audio de la sesión de codificación y los almacena en un buffer.
func (ad *AudioDownloader) readAudioFramesToBuffer(session encoder.EncodeSession) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	for {
		frame, err := session.ReadFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error al leer frame de audio: %w", err)
		}

		_, err = buffer.Write(frame)
		if err != nil {
			return nil, fmt.Errorf("error al escribir frame en buffer: %w", err)
		}
	}
	return &buffer, nil
}
