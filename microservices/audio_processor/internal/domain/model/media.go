package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

type (

	// Media representa un modelo de procesamiento multimedia.
	Media struct {
		ID      string `json:"id" bson:"_id" dynamodbav:"PK"`
		VideoID string `json:"video_id" bson:"video_id" dynamodbav:"SK"`
		// Title es el título de la canción.
		// Representa el nombre de la canción tal como aparece en la fuente de origen.
		Title          string            `json:"title" bson:"title" dynamodbav:"title"`
		Status         string            `json:"status" bson:"status" dynamodbav:"status"`
		Message        string            `json:"message" bson:"message" dynamodbav:"message"`
		Metadata       *PlatformMetadata `json:"metadata" bson:"metadata" dynamodbav:"metadata"`
		FileData       *FileData         `json:"file_data" bson:"file_data" dynamodbav:"file_data"`
		ProcessingDate time.Time         `json:"processing_date" bson:"processing_date" dynamodbav:"processing_date"`
		Success        bool              `json:"success" bson:"success" dynamodbav:"success"`
		Attempts       int               `json:"attempts" bson:"attempts" dynamodbav:"attempts"`
		Failures       int               `json:"failures" bson:"failures" dynamodbav:"failures"`
		CreatedAt      time.Time         `json:"created_at" bson:"created_at" dynamodbav:"created_at"`
		UpdatedAt      time.Time         `json:"updated_at" bson:"updated_at" dynamodbav:"updated_at"`
		PlayCount      int               `json:"play_count" bson:"play_count" dynamodbav:"play_count"`
	}

	// PlatformMetadata representa los metadatos de una plataforma
	PlatformMetadata struct {
		// DurationMs es la duración de la canción en milisegundos.
		// Representa cuánto tiempo dura la canción desde el inicio hasta el final.
		DurationMs int64 `json:"duration_ms" bson:"duration_ms" dynamodbav:"duration_ms"`
		// es la URL de la canción en YouTube.
		// Permite localizar la canción en YouTube para referencias adicionales o reproducción.
		URL string `json:"url" bson:"url" dynamodbav:"url"`
		// ThumbnailURL es la imagen en miniatura del contenido.
		// Proporciona una vista previa visual de la canción, útil para interfaces de usuario y presentación.
		ThumbnailURL string `json:"thumbnail_url" bson:"thumbnail_url" dynamodbav:"thumbnail_url"`
		// Platform indica la plataforma de origen de la canción (e.g., YouTube).
		// Este campo identifica la fuente desde la cual se obtuvo la canción.
		Platform string `json:"platform" bson:"platform" dynamodbav:"platform"`
	}

	// FileData contiene información sobre el archivo de la canción procesada.
	// Esto incluye la ruta del archivo, el tamaño del archivo, el tipo de archivo
	// y la URL pública del archivo.
	FileData struct {
		// FilePath es la ruta del archivo de la canción procesada.
		FilePath string `bson:"file_path" json:"file_path" dynamodbav:"file_path"`

		// FileSize es el tamaño del archivo de la canción procesada.
		FileSize string `bson:"file_size" json:"file_size" dynamodbav:"file_size"`

		// FileType es el tipo de archivo de la canción procesada.
		FileType string `bson:"file_type" json:"file_type" dynamodbav:"file_type"`
	}

	MediaDetails struct {
		Title        string
		ID           string
		Description  string
		Creator      string
		DurationMs   int64
		PublishedAt  time.Time
		URL          string
		ThumbnailURL string
		Provider     string
	}
)

func (f *FileData) ToAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"file_path": &types.AttributeValueMemberS{Value: f.FilePath},
			"file_size": &types.AttributeValueMemberS{Value: f.FileSize},
			"file_type": &types.AttributeValueMemberS{Value: f.FileType},
		},
	}
}

func (m *PlatformMetadata) ToAttributeValue() types.AttributeValue {
	return &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"duration_ms":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", m.DurationMs)},
			"url":           &types.AttributeValueMemberS{Value: m.URL},
			"thumbnail_url": &types.AttributeValueMemberS{Value: m.ThumbnailURL},
			"platform":      &types.AttributeValueMemberS{Value: m.Platform},
		},
	}
}
