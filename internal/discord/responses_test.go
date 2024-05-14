package discord

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

// MockSessionService es una implementación de SessionService para usar en pruebas.
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) InteractionRespond(i *discordgo.Interaction, r *discordgo.InteractionResponse) error {
	args := m.Called(i, r)
	return args.Error(0)
}

func (m *MockSessionService) FollowupMessageCreate(i *discordgo.Interaction, wait bool, params *discordgo.WebhookParams) (*discordgo.Message, error) {
	args := m.Called(i, wait, params)
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

func TestDiscordResponseHandler_Respond(t *testing.T) {
	logger := zap.NewNop()
	responseHandler := NewDiscordResponseHandler(logger)
	mockSession := new(MockSessionService)

	interaction := &discordgo.Interaction{}
	response := discordgo.InteractionResponse{}

	mockSession.On("InteractionRespond", interaction, &response).Return(nil)

	err := responseHandler.Respond(mockSession, interaction, response)
	if err != nil {
		t.Errorf("Se esperaba error nulo, pero se obtuvo: %v", err)
	}

	mockSession.AssertExpectations(t)
}

func TestDiscordResponseHandler_Respond_Error(t *testing.T) {
	logger := zap.NewNop()
	responseHandler := NewDiscordResponseHandler(logger)
	mockSession := new(MockSessionService)

	interaction := &discordgo.Interaction{}
	response := discordgo.InteractionResponse{}

	mockSession.On("InteractionRespond", interaction, &response).Return(errors.New("error al responder"))

	err := responseHandler.Respond(mockSession, interaction, response)
	if err == nil {
		t.Error("Se esperaba un error, pero no se recibió ninguno")
	}

	// Verifica que el metodo del mock haya sido llamado exactamente una vez
	mockSession.AssertNumberOfCalls(t, "InteractionRespond", 1)

	// Verifica el mensaje de error retornado por la funcion
	expectedError := "error al responder"
	if err.Error() != expectedError {
		t.Errorf("Se esperaba un error '%s', pero se obtuvo: '%s'", expectedError, err.Error())
	}

	mockSession.AssertExpectations(t)
}

func TestDiscordResponseHandler_RespondWithMessage(t *testing.T) {
	logger := zap.NewNop()
	responseHandler := NewDiscordResponseHandler(logger)
	mockSession := new(MockSessionService)

	interaction := &discordgo.Interaction{}
	message := "¡Hola desde las pruebas!"
	expectedResponse := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}

	mockSession.On("InteractionRespond", interaction, &expectedResponse).Return(nil)

	err := responseHandler.RespondWithMessage(mockSession, interaction, message)
	if err != nil {
		t.Errorf("Se esperaba error nulo, pero se obtuvo: %v", err)
	}

	// Verifica que el metodo del mock haya sido llamado exactamente una vez
	mockSession.AssertNumberOfCalls(t, "InteractionRespond", 1)

	mockSession.AssertExpectations(t)
}

func TestDiscordResponseHandler_CreateFollowupMessage(t *testing.T) {
	logger := zap.NewNop()
	responseHandler := NewDiscordResponseHandler(logger)
	mockSession := new(MockSessionService)

	interaction := &discordgo.Interaction{}
	params := discordgo.WebhookParams{}
	expectedMessage := &discordgo.Message{}

	mockSession.On("FollowupMessageCreate", interaction, true, &params).Return(expectedMessage, nil)

	err := responseHandler.CreateFollowupMessage(mockSession, interaction, params)
	if err != nil {
		t.Errorf("Se esperaba error nulo, pero se obtuvo: %v", err)
	}

	mockSession.AssertExpectations(t)
}

func TestDiscordResponseHandler_CreateFollowupMessage_Error(t *testing.T) {
	logger := zap.NewNop()
	responseHandler := NewDiscordResponseHandler(logger)
	mockSession := new(MockSessionService)

	interaction := &discordgo.Interaction{}
	params := discordgo.WebhookParams{}

	mockSession.On("FollowupMessageCreate", interaction, true, &params).Return(&discordgo.Message{}, errors.New("error al crear mensaje de seguimiento"))

	err := responseHandler.CreateFollowupMessage(mockSession, interaction, params)
	if err == nil {
		t.Error("Se esperaba un error, pero no se recibió ninguno")
	}

	// Verifica que el metodo del mock haya sido llamado exactamente una vez
	mockSession.AssertNumberOfCalls(t, "FollowupMessageCreate", 1)

	// Verifica que el error retornado por la funcinn es el esperado
	expectedError := "error al crear mensaje de seguimiento"
	if err.Error() != expectedError {
		t.Errorf("Se esperaba un error '%s', pero se obtuvo: '%s'", expectedError, err.Error())
	}

	mockSession.AssertExpectations(t)
}
