package model

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Metadata representa la información sobre una canción procesada.
// Contiene detalles sobre la canción, como su título, artista, duración, y URLs de recursos.
type Metadata struct {
	// ID es un identificador único para la canción.
	// Este campo se utiliza para asociar la metadata con una canción específica.
	ID string `bson:"_id" json:"id" dynamodbav:"id"`

	// VideoID Es el identificador del video
	VideoID string `bson:"video_id" json:"video_id" dynamodbav:"video_id"`

	// Title es el título de la canción.
	// Representa el nombre de la canción tal como aparece en la fuente de origen.
	Title string `bson:"title" json:"title" dynamodbav:"title"`

	GSI2PK string `bson:"-" json:"-" dynamodbav:"GSI2_PK"`

	// DurationMs es la duración de la canción en milisegundos.
	// Representa cuánto tiempo dura la canción desde el inicio hasta el final.
	DurationMs int64 `bson:"duration_ms" json:"duration_ms" dynamodbav:"duration_ms"`

	// es la URL de la canción en YouTube.
	// Permite localizar la canción en YouTube para referencias adicionales o reproducción.
	URL string `bson:"url" json:"url" dynamodbav:"url"`

	// ThumbnailURL es la imagen en miniatura del contenido.
	// Proporciona una vista previa visual de la canción, útil para interfaces de usuario y presentación.
	ThumbnailURL string `bson:"thumbnail_url" json:"thumbnail_url" dynamodbav:"thumbnail_url"`

	// Platform indica la plataforma de origen de la canción (e.g., YouTube).
	// Este campo identifica la fuente desde la cual se obtuvo la canción.
	Platform string `bson:"platform" json:"platform" dynamodbav:"platform"`
}

func (m *Metadata) ToAttributeValue() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: m.ID},
		"video_id":    &types.AttributeValueMemberS{Value: m.VideoID},
		"title":       &types.AttributeValueMemberS{Value: m.Title},
		"duration_ms": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", m.DurationMs)},
		"url":         &types.AttributeValueMemberS{Value: m.URL},
		"thumbnail":   &types.AttributeValueMemberS{Value: m.ThumbnailURL},
		"platform":    &types.AttributeValueMemberS{Value: m.Platform},
	}
}
