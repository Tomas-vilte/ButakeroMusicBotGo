package model

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

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

	// Duration es la duración de la canción en segundos.
	// Representa cuánto tiempo dura la canción desde el inicio hasta el final.
	Duration string `bson:"duration" json:"duration" dynamodbav:"duration"`

	// URLYouTube es la URL de la canción en YouTube.
	// Permite localizar la canción en YouTube para referencias adicionales o reproducción.
	URLYouTube string `bson:"url_youtube" json:"url_youtube" dynamodbav:"url_youtube"`

	// Thumbnail es la imagen en miniatura del contenido.
	// Proporciona una vista previa visual de la canción, útil para interfaces de usuario y presentación.
	Thumbnail string `bson:"thumbnail" json:"thumbnail" dynamodbav:"thumbnail"`

	// Platform indica la plataforma de origen de la canción (e.g., YouTube).
	// Este campo identifica la fuente desde la cual se obtuvo la canción.
	Platform string `bson:"platform" json:"platform" dynamodbav:"platform"`
}

func (m *Metadata) ToAttributeValue() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: m.ID},
		"video_id":    &types.AttributeValueMemberS{Value: m.VideoID},
		"title":       &types.AttributeValueMemberS{Value: m.Title},
		"duration":    &types.AttributeValueMemberS{Value: m.Duration},
		"url_youtube": &types.AttributeValueMemberS{Value: m.URLYouTube},
		"thumbnail":   &types.AttributeValueMemberS{Value: m.Thumbnail},
		"platform":    &types.AttributeValueMemberS{Value: m.Platform},
	}
}
