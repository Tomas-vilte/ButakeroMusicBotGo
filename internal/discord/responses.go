package discord

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// SessionService define la interfaz para los métodos de discordgo.Session que necesitamos.
type SessionService interface {
	InteractionRespond(i *discordgo.Interaction, r *discordgo.InteractionResponse) error
	FollowupMessageCreate(i *discordgo.Interaction, wait bool, params *discordgo.WebhookParams) (*discordgo.Message, error)
}

// DiscordSessionService es una implementación real de SessionService que envuelve discordgo.Session.
type DiscordSessionService struct {
	session *discordgo.Session
}

func NewSessionService(session *discordgo.Session) *DiscordSessionService {
	return &DiscordSessionService{
		session: session,
	}
}

func (s *DiscordSessionService) InteractionRespond(i *discordgo.Interaction, r *discordgo.InteractionResponse) error {
	return s.session.InteractionRespond(i, r)
}

func (s *DiscordSessionService) FollowupMessageCreate(i *discordgo.Interaction, wait bool, params *discordgo.WebhookParams) (*discordgo.Message, error) {
	return s.session.FollowupMessageCreate(i, wait, params)
}

// ResponseHandler define la interfaz para manejar respuestas a interacciones de Discord.
type ResponseHandler interface {
	Respond(session SessionService, interaction *discordgo.Interaction, response discordgo.InteractionResponse) error
	RespondWithMessage(session SessionService, interaction *discordgo.Interaction, message string) error
	CreateFollowupMessage(session SessionService, interaction *discordgo.Interaction, params discordgo.WebhookParams) error
}

// DiscordResponseHandler implementa la interfaz ResponseHandler para manejar respuestas a interacciones de Discord.
type DiscordResponseHandler struct {
	logger *zap.Logger
}

// NewDiscordResponseHandler crea una nueva instancia de DiscordResponseHandler.
func NewDiscordResponseHandler(logger *zap.Logger) *DiscordResponseHandler {
	return &DiscordResponseHandler{logger: logger}
}

// Respond responde a una interacción de Discord con la respuesta proporcionada.
func (h *DiscordResponseHandler) Respond(session SessionService, interaction *discordgo.Interaction, response discordgo.InteractionResponse) error {
	if err := session.InteractionRespond(interaction, &response); err != nil {
		h.logger.Error("No se pudo responder a la interacción", zap.Error(err))
		return err
	}
	return nil
}

// RespondWithMessage responde a una interacción de Discord con un mensaje de texto.
func (h *DiscordResponseHandler) RespondWithMessage(session SessionService, interaction *discordgo.Interaction, message string) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}
	return h.Respond(session, interaction, response)
}

// CreateFollowupMessage crea un mensaje de seguimiento para una interacción de Discord.
func (h *DiscordResponseHandler) CreateFollowupMessage(session SessionService, interaction *discordgo.Interaction, params discordgo.WebhookParams) error {
	if _, err := session.FollowupMessageCreate(interaction, true, &params); err != nil {
		h.logger.Error("No se pudo crear el mensaje de seguimiento", zap.Error(err))
		return err
	}
	return nil
}
