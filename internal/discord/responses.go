package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"log"
)

// InteractionRespond responde a una interacción con la respuesta proporcionada.
func InteractionRespond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) {
	if err := s.InteractionRespond(i, response); err != nil {
		log.Printf("falló al responder a la interacción: %v", err)
	}
}

// InteractionRespondServerError responde a una interacción con un mensaje de error de servidor.
func InteractionRespondServerError(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction) {
	InteractionRespond(logger, s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "seso tiene algunos problemas...",
		},
	})
}

// InteractionRespondMessage responde a una interacción con un mensaje proporcionado.
func InteractionRespondMessage(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, message string) {
	InteractionRespond(logger, s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

// FollowupMessageCreate crea un mensaje de seguimiento para una interacción.
func FollowupMessageCreate(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, params *discordgo.WebhookParams) {
	if _, err := s.FollowupMessageCreate(i, true, params); err != nil {
		logger.Error("falló al crear el mensaje de seguimiento", zap.Error(err))
	}
}
