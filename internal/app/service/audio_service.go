package service

import (
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"log"
)

// AudioService define los métodos para interactuar con la funcionalidad de audio.
type AudioService interface {
	PlayAudio(session *discordgo.Session, message *discordgo.MessageCreate, audioURL string) error
}

// audioService implementa AudioService.
type audioService struct{}

// NewAudioService crea una nueva instancia de AudioService.
func NewAudioService() AudioService {
	return &audioService{}
}

// PlayAudio reproduce el audio en el canal de voz del autor del mensaje.
func (s *audioService) PlayAudio(session *discordgo.Session, message *discordgo.MessageCreate, audioURL string) error {
	// Obtener el ID del canal de voz del autor del mensaje
	channelID, err := getAuthorVoiceChannelID(session, message)
	if err != nil {
		log.Printf("Error al obtener el ID del canal de voz del autor: %v", err)
		return err
	}

	// Unirse al canal de voz
	conn, err := session.ChannelVoiceJoin(message.GuildID, channelID, false, true)
	if err != nil {
		log.Printf("Error al unirse al canal de voz: %v", err)
		return err
	}
	defer conn.Close()

	// Reproducir el audio en segundo plano
	log.Printf("Reproduciendo audio en el canal de voz del autor")
	go dgvoice.PlayAudioFile(conn, audioURL, make(chan bool))

	return nil
}

// Función para obtener el ID del canal de voz del autor del mensaje
func getAuthorVoiceChannelID(session *discordgo.Session, message *discordgo.MessageCreate) (string, error) {
	// Obtener el estado de Discord para el servidor (guild) del mensaje
	guild, err := session.State.Guild(message.GuildID)
	if err != nil {
		log.Printf("Error al obtener el servidor (guild): %v", err)
		return "", fmt.Errorf("error al obtener el servidor (guild): %w", err)
	}

	// Buscar al autor del mensaje en la lista de miembros del servidor (guild)
	for _, vs := range guild.VoiceStates {
		if vs.UserID == message.Author.ID {
			// Si el autor está en un canal de voz, devolvemos el ID del canal
			if vs.ChannelID != "" {
				log.Printf("Usuario %s encontrado en el canal de voz %s", message.Author.Username, vs.ChannelID)
				return vs.ChannelID, nil
			}
			// Si el autor no está en un canal de voz, devolvemos un error
			log.Printf("El usuario %s no está en un canal de voz", message.Author.Username)
			return "", fmt.Errorf("el usuario %s no está en un canal de voz", message.Author.Username)
		}
	}

	// Si el autor del mensaje no está en la lista de miembros del servidor (guild), devolvemos un error
	log.Printf("Usuario %s no encontrado en el estado del servidor (guild)", message.Author.Username)
	return "", fmt.Errorf("el usuario %s no se encuentra en el estado del servidor (guild)", message.Author.Username)
}
