package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type InteractionResponder interface {
	Respond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse)
}

type InteractionResponse struct {
	Message string
	Type    discordgo.InteractionResponseType
}

func (ir *InteractionResponse) Respond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction) {
	InteractionRespond(logger, s, i, &discordgo.InteractionResponse{
		Type: ir.Type,
		Data: &discordgo.InteractionResponseData{
			Content: ir.Message,
		},
	})
}

func InteractionRespond(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, d *discordgo.InteractionResponse) {
	if err := s.InteractionRespond(i, d); err != nil {
		logger.Error("no pudo responder a la interacción", zap.Error(err))
	}
}

func FollowupMessageCreator(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction) *discordgo.WebhookParams {
	return &discordgo.WebhookParams{
		Content: "hay algunos problemas...",
	}
}

func FollowupMessageCreate(logger *zap.Logger, s *discordgo.Session, i *discordgo.Interaction, params *discordgo.WebhookParams) {
	if _, err := s.FollowupMessageCreate(i, true, params); err != nil {
		logger.Error("no se pudo crear el mensaje de seguimiento", zap.Error(err))
	}
}
