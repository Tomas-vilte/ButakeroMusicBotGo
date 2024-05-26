package messaging

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestSendMessage(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	client := NewDiscordGoClient(mockSession, mockLogger)

	channelID := "1234567890"
	embed := &discordgo.MessageEmbed{Title: "Test"}
	mockSession.On("ChannelMessageSendEmbed", channelID, embed, mock.Anything).Return(&discordgo.Message{}, nil)

	// Configurar el mock del logger
	mockLogger.On("Info", "Mensaje enviado al canal", mock.Anything).Return()

	err := client.SendMessage(channelID, embed)

	assert.NoError(t, err)

	mockSession.AssertCalled(t, "ChannelMessageSendEmbed", channelID, embed, mock.Anything)
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado al canal", mock.Anything)
}

func TestSendMessage_ErrorSendingMessage(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	client := NewDiscordGoClient(mockSession, mockLogger)

	channelID := "1234567890"
	embed := &discordgo.MessageEmbed{Title: "Test"}
	expectedError := errors.New("error al enviar el mensaje")
	mockSession.On("ChannelMessageSendEmbed", channelID, embed, mock.Anything).Return(&discordgo.Message{}, expectedError)

	mockLogger.On("Error", "Error al enviar mensaje al canal", []zapcore.Field{
		zap.String("ID del canal:", channelID), zap.Error(expectedError),
	}).Return()

	err := client.SendMessage(channelID, embed)

	assert.EqualError(t, err, expectedError.Error())

	mockSession.AssertCalled(t, "ChannelMessageSendEmbed", channelID, embed, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al enviar mensaje al canal", []zapcore.Field{
		zap.String("ID del canal:", channelID), zap.Error(expectedError),
	})
}

func TestSendMessageToServers_ErrorGettingGuilds(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)
	guilds := []*discordgo.UserGuild{{ID: "123456", Name: "Test Guild"}}
	client := NewDiscordGoClient(mockSession, mockLogger)

	expectedError := errors.New("error al obtener la lista de servidores")
	mockSession.On("UserGuilds", 0, "", "", true, mock.Anything).Return(guilds, expectedError)

	mockLogger.On("Error", "Error al obtener la lista de servidores", []zapcore.Field{
		zap.Error(expectedError),
	}).Return()

	embed := &discordgo.MessageEmbed{Title: "Test"}
	err := client.SendMessageToServers(embed)

	assert.EqualError(t, err, expectedError.Error())

	mockSession.AssertCalled(t, "UserGuilds", 0, "", "", true, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al obtener la lista de servidores", []zapcore.Field{
		zap.Error(expectedError),
	})
}

func TestSendMessageToServers_ErrorFindingStatusBotChannel(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	client := NewDiscordGoClient(mockSession, mockLogger)

	guildID := "guild1"
	guilds := []*discordgo.UserGuild{{ID: guildID, Name: "Test Guild"}}
	mockSession.On("UserGuilds", 0, "", "", true, mock.Anything).Return(guilds, nil)

	expectedError := errors.New("error al buscar el canal")
	mockSession.On("GuildChannels", guildID, mock.Anything).Return([]*discordgo.Channel{}, expectedError)

	mockLogger.On("Error", "Error al verificar el canal 'statusBot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al obtener los channels", mock.AnythingOfType("[]zapcore.Field")).Return()

	embed := &discordgo.MessageEmbed{Title: "Test"}
	err := client.SendMessageToServers(embed)

	assert.NoError(t, err)

	mockSession.AssertCalled(t, "UserGuilds", 0, "", "", true, mock.Anything)
	mockSession.AssertCalled(t, "GuildChannels", guildID, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al verificar el canal 'statusBot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Error", "Error al obtener los channels", mock.AnythingOfType("[]zapcore.Field"))
}

func TestDiscordGoClient_SendMessageToServers_CreateChannel(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	guild := &discordgo.UserGuild{
		ID:   "123",
		Name: "Test Guild",
	}
	expectedChannel := &discordgo.Channel{
		ID:   "456",
		Name: "status-bot",
	}
	mockSession.On("UserGuilds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*discordgo.UserGuild{guild}, nil)
	mockSession.On("GuildChannels", guild.ID, mock.Anything).Return([]*discordgo.Channel{}, nil)
	mockSession.On("GuildChannelCreate", guild.ID, "Status Bot", discordgo.ChannelTypeGuildCategory, mock.Anything).Return(&discordgo.Channel{
		ID:   "789",
		Name: "Status Bot",
		Type: discordgo.ChannelTypeGuildCategory,
	}, nil)
	mockSession.On("GuildChannelCreateComplex", guild.ID, mock.Anything, mock.Anything).Return(expectedChannel, nil)
	mockSession.On("ChannelMessageSendEmbed", expectedChannel.ID, mock.Anything, mock.Anything).Return(&discordgo.Message{}, nil)
	mockLogger.On("Info", "Categoría 'Status Bot' creada en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Canal 'statusBot' creado en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Mensaje enviado al canal", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Mensaje enviado a Discord en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	client := NewDiscordGoClient(mockSession, mockLogger)

	err := client.SendMessageToServers(&discordgo.MessageEmbed{})

	assert.NoError(t, err)

	mockSession.AssertExpectations(t)
	mockLogger.AssertCalled(t, "Info", "Canal 'statusBot' creado en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Info", "Categoría 'Status Bot' creada en el servidor", mock.AnythingOfType("[]zapcore.Field"))
	mockSession.AssertCalled(t, "ChannelMessageSendEmbed", expectedChannel.ID, mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado al canal", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado a Discord en el servidor", mock.AnythingOfType("[]zapcore.Field"))
}

func TestDiscordGoClient_SendMessageToServers_SendMessage(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	guild := &discordgo.UserGuild{
		ID:   "123",
		Name: "Test Guild",
	}
	channel := &discordgo.Channel{
		ID:   "456",
		Name: "status-bot",
	}
	embed := &discordgo.MessageEmbed{
		Title: "Test Embed",
	}

	mockSession.On("UserGuilds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*discordgo.UserGuild{guild}, nil)
	mockSession.On("GuildChannels", guild.ID, mock.Anything).Return([]*discordgo.Channel{channel}, nil)
	mockSession.On("ChannelMessageSendEmbed", channel.ID, embed, mock.Anything).Return(&discordgo.Message{}, nil)

	mockLogger.On("Info", "Mensaje enviado al canal", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Mensaje enviado a Discord en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()

	client := NewDiscordGoClient(mockSession, mockLogger)
	err := client.SendMessageToServers(embed)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado a Discord en el servidor", mock.Anything)
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado al canal", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Info", "Mensaje enviado a Discord en el servidor", mock.AnythingOfType("[]zapcore.Field"))

	mockSession.ExpectedCalls = nil
	mockLogger.ExpectedCalls = nil

	mockSession.On("UserGuilds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*discordgo.UserGuild{guild}, nil)
	mockSession.On("GuildChannels", guild.ID, mock.Anything).Return([]*discordgo.Channel{channel}, nil)
	mockSession.On("ChannelMessageSendEmbed", channel.ID, embed, mock.Anything).Return(&discordgo.Message{}, errors.New("error al enviar mensaje"))

	mockLogger.On("Error", "Error al enviar el mensaje a Discord en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al enviar mensaje al canal", mock.AnythingOfType("[]zapcore.Field")).Return()

	client = NewDiscordGoClient(mockSession, mockLogger)
	err = client.SendMessageToServers(embed)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Error", "Error al enviar el mensaje a Discord en el servidor", mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al enviar mensaje al canal", mock.AnythingOfType("[]zapcore.Field"))
}

func TestCreateStatusBotChannel_ErrorCreatingCategory(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)

	client := NewDiscordGoClient(mockSession, mockLogger)

	guildID := "guild1"
	expectedError := errors.New("error al crear la categoria")
	mockSession.On("GuildChannels", guildID, mock.Anything).Return([]*discordgo.Channel{}, nil)
	mockSession.On("GuildChannelCreate", guildID, "Status Bot", discordgo.ChannelTypeGuildCategory, mock.Anything).Return(&discordgo.Channel{}, expectedError)

	mockLogger.On("Error", "Error al crear la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()

	channel, err := client.createStatusBotChannel(guildID)

	assert.Nil(t, channel)
	assert.EqualError(t, err, expectedError.Error())

	mockSession.AssertCalled(t, "GuildChannelCreate", guildID, "Status Bot", discordgo.ChannelTypeGuildCategory, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al crear la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))
}

func TestCreateStatusBotCategory_ErrorFetchingChannels(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)
	client := NewDiscordGoClient(mockSession, mockLogger)

	guildID := "guild1"
	expectedError := errors.New("error al obtener los canales")
	mockSession.On("GuildChannels", guildID, mock.Anything).Return([]*discordgo.Channel{}, expectedError)

	mockLogger.On("Error", "Error al obtener los channels", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al crear la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()

	channel, err := client.createStatusBotChannel(guildID)

	assert.Nil(t, channel)
	assert.EqualError(t, err, expectedError.Error())

	mockSession.AssertCalled(t, "GuildChannels", guildID, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al obtener los channels", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Error", "Error al crear la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))

}

func TestCreateStatusBotCategory_CategoryAlreadyExists(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)
	client := NewDiscordGoClient(mockSession, mockLogger)

	guildID := "guild1"
	existingCategory := &discordgo.Channel{Name: "Status Bot", Type: discordgo.ChannelTypeGuildCategory}
	guildChannels := []*discordgo.Channel{existingCategory}
	mockSession.On("GuildChannels", guildID, mock.Anything).Return(guildChannels, nil)

	mockLogger.On("Error", "La categoria ya existe en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()

	category, err := client.createStatusBotCategory(guildID)
	assert.NotNil(t, category)
	assert.NoError(t, err)
	assert.Equal(t, existingCategory, category)

	mockSession.AssertCalled(t, "GuildChannels", guildID, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "La categoria ya existe en el servidor", mock.AnythingOfType("[]zapcore.Field"))

}

func TestCreateStatusBotChannel_ErrorCreatingChannel(t *testing.T) {
	mockSession := new(MockDiscordSession)
	mockLogger := new(MockLogger)
	client := NewDiscordGoClient(mockSession, mockLogger)
	guildID := "guild1"
	category := &discordgo.Channel{ID: "category1", Name: "Status Bot", Type: discordgo.ChannelTypeGuildCategory}

	mockSession.On("GuildChannels", guildID, mock.Anything).Return([]*discordgo.Channel{}, nil)
	mockSession.On("GuildChannelCreate", guildID, "Status Bot", discordgo.ChannelTypeGuildCategory, mock.Anything).Return(category, nil)

	expectedError := errors.New("error al crear el canal")
	mockSession.On("GuildChannelCreateComplex", guildID, discordgo.GuildChannelCreateData{
		Name:     "status-bot",
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: category.ID,
	}, mock.Anything).Return(&discordgo.Channel{}, expectedError)

	mockLogger.On("Error", "Error al crear el canal 'status-bot' en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Info", "Categoría 'Status Bot' creada en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()
	mockLogger.On("Error", "Error al crear el canal 'status-bot' en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field")).Return()

	channel, err := client.createStatusBotChannel(guildID)

	assert.Nil(t, channel)
	assert.EqualError(t, err, expectedError.Error())

	mockSession.AssertCalled(t, "GuildChannelCreateComplex", guildID, discordgo.GuildChannelCreateData{
		Name:     "status-bot",
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: category.ID,
	}, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al crear el canal 'status-bot' en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Info", "Categoría 'Status Bot' creada en el servidor", mock.AnythingOfType("[]zapcore.Field"))
	mockLogger.AssertCalled(t, "Error", "Error al crear el canal 'status-bot' en la categoría 'Status Bot' en el servidor", mock.AnythingOfType("[]zapcore.Field"))
}
