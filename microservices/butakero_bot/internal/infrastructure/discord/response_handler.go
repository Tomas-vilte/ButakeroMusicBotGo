package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type ResponseHandler struct {
	session *discordgo.Session
	logging logging.Logger
}

func NewResponseHandler(session *discordgo.Session, logging logging.Logger) *ResponseHandler {
	return &ResponseHandler{
		session: session,
		logging: logging,
	}
}

// Respond Responde a una interacción de Discord.
func (h *ResponseHandler) Respond(interaction *discordgo.Interaction, response discordgo.InteractionResponse) error {
	if err := h.session.InteractionRespond(interaction, &response); err != nil {
		h.logging.Error("No se pudo responder a la interacción",
			zap.String("interaction_id", interaction.ID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// RespondWithMessage Responde con un mensaje de texto simple.
func (h *ResponseHandler) RespondWithMessage(interaction *discordgo.Interaction, message string) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}
	return h.Respond(interaction, response)
}

// CreateFollowupMessage Crea un mensaje de seguimiento después de una interacción.
func (h *ResponseHandler) CreateFollowupMessage(interaction *discordgo.Interaction, params discordgo.WebhookParams) error {
	if _, err := h.session.FollowupMessageCreate(interaction, true, &params); err != nil {
		h.logging.Error("No se pudo crear un mensaje de seguimiento", zap.Error(err))
		return err
	}
	return nil
}
