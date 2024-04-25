package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// InteractionResponder es responsable de responder a las interacciones.
type InteractionResponder interface {
	Respond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction)
	RespondServerError(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction)
	InteractionRespondMessage(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, message string)
}

type InteractionResponse struct {
	Message string
	Type    discordgo.InteractionResponseType
}

func (ir *InteractionResponse) Respond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction) {
	response := &discordgo.InteractionResponse{
		Type: ir.Type,
		Data: &discordgo.InteractionResponseData{
			Content: ir.Message,
		},
	}
	RespondToInteraction(logger, s, i, response)
}

func RespondToInteraction(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) {
	if err := s.InteractionRespond(i, response); err != nil {
		logger.Error("no pudo responder a la interacción", zap.Error(err))
	}
}

func (ir *InteractionResponse) RespondServerError(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction) {
	RespondToInteraction(logger, s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Hubo problemas...",
		},
	})
}

func (ir *InteractionResponse) InteractionRespondMessage(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, message string) {
	RespondToInteraction(logger, s, i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func FollowupMessageCreate(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, params *discordgo.WebhookParams) {
	if _, err := s.FollowupMessageCreate(i, true, params); err != nil {
		logger.Error("failed to create followup message", zap.Error(err))
	}
}
