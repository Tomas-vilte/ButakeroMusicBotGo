package queuing

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
)

type MockEventFormatter struct {
	mock.Mock
}

func (m *MockEventFormatter) FormatEvent(event map[string]interface{}) (*discordgo.MessageEmbed, error) {
	args := m.Called(event)
	return args.Get(0).(*discordgo.MessageEmbed), args.Error(1)
}
