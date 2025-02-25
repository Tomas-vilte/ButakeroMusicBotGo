package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/bwmarrin/discordgo"
)

type InteractionHandler struct {
	players          map[string]*MusicPlayer // Mapa guildID -> MusicPlayer
	interactionStore ports.InteractionStorage
	discordMessenger ports.DiscordMessenger
	mu               sync.Mutex

	// Factories para crear dependencias específicas por servidor
	voiceSessionFactory func(guildID string) ports.VoiceSession
	decoderFactory      func(io.ReadCloser) ports.Decoder

	// Otras dependencias compartidas
	messageConsumer     ports.MessageConsumer
	externalSongService ports.ExternalSongService
	songRepo            ports.SongRepository
	songStorage         ports.SongStorage
	stateStorage        ports.StateStorage
	storageAudio        ports.StorageAudio
}

func NewInteractionHandler(
	interactionStore ports.InteractionStorage,
	discordMessenger ports.DiscordMessenger,
	voiceSessionFactory func(guildID string) ports.VoiceSession,
	decoderFactory func(io.ReadCloser) ports.Decoder,
	messageConsumer ports.MessageConsumer,
	externalSongService ports.ExternalSongService,
	songRepo ports.SongRepository,
	songStorage ports.SongStorage,
	stateStorage ports.StateStorage,
	storageAudio ports.StorageAudio,
) *InteractionHandler {
	return &InteractionHandler{
		players:             make(map[string]*MusicPlayer),
		interactionStore:    interactionStore,
		discordMessenger:    discordMessenger,
		voiceSessionFactory: voiceSessionFactory,
		decoderFactory:      decoderFactory,
		messageConsumer:     messageConsumer,
		externalSongService: externalSongService,
		songRepo:            songRepo,
		songStorage:         songStorage,
		stateStorage:        stateStorage,
		storageAudio:        storageAudio,
	}
}

// HandleInteraction procesa los eventos de interacción
func (h *InteractionHandler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommandInteraction(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponentInteraction(s, i)
	}
}

// handleCommandInteraction procesa los comandos de barra
func (h *InteractionHandler) handleCommandInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "play":
		h.handlePlayCommand(s, i)
	case "stop":
		h.handleStopCommand(s, i)
	case "skip":
		h.handleSkipCommand(s, i)
	case "queue":
		h.handleQueueCommand(s, i)
	default:
		h.sendResponse(s, i, "Comando no reconocido")
	}
}

// handlePlayCommand procesa el comando /play
func (h *InteractionHandler) handlePlayCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Indicar a Discord que estamos procesando la interacción
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		//h.logger.Error("Error al responder a la interacción", zap.Error(err))
		return
	}

	// Obtener los datos de la interacción
	data := i.ApplicationCommandData()
	options := data.Options

	if len(options) == 0 {
		// Enviar un mensaje de error
		_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Debes proporcionar una URL o nombre de canción",
		})
		return
	}

	input := options[0].StringValue()
	guildID := i.GuildID
	channelID := i.ChannelID

	// Llamar al método Play del reproductor en una goroutine
	go func() {
		player := h.getOrCreatePlayer(guildID)
		err := player.Play(context.Background(), channelID, input)
		if err != nil {
			// Enviar un mensaje de error
			_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Error al reproducir: %v", err),
			})
			return
		}

		// Enviar un mensaje de éxito
		_, _ = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Canción agregada a la lista de reproducción",
		})
	}()
}

// handleStopCommand procesa el comando /stop
func (h *InteractionHandler) handleStopCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	player := h.getOrCreatePlayer(guildID)

	// Detener la reproducción
	err := player.Stop()
	if err != nil {
		_ = h.discordMessenger.SendText(i.ChannelID, fmt.Sprintf("Error al detener: %v", err))
		return
	}

	// Limpiar la lista de reproducción
	h.interactionStore.DeleteSongList(i.ChannelID)

	_ = h.discordMessenger.SendText(i.ChannelID, "Reproducción detenida y lista de reproducción limpiada")
}

// handleSkipCommand procesa el comando /skip
func (h *InteractionHandler) handleSkipCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	player := h.getOrCreatePlayer(guildID)

	// Llamar al método Skip del reproductor
	err := player.Skip()
	if err != nil {
		h.sendResponse(s, i, fmt.Sprintf("Error al saltar: %v", err))
		return
	}

	h.sendResponse(s, i, "Skipping to the next song")
}

// handleQueueCommand procesa el comando /queue
func (h *InteractionHandler) handleQueueCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID

	// Obtener la lista de reproducción
	songList := h.interactionStore.GetSongList(channelID)
	if len(songList) == 0 {
		_ = h.discordMessenger.SendText(channelID, "La lista de reproducción está vacía")
		return
	}

	// Construir el mensaje con la lista de reproducción
	var queueMsg strings.Builder
	queueMsg.WriteString("**Lista de reproducción:**\n")
	for idx, song := range songList {
		queueMsg.WriteString(fmt.Sprintf("%d. %s\n", idx+1, song.Title))
	}

	_ = h.discordMessenger.SendText(channelID, queueMsg.String())
}

// handleComponentInteraction procesa interacciones con componentes (botones, menús, etc.)
func (h *InteractionHandler) handleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Aquí puedes manejar interacciones con botones o menús
	h.sendResponse(s, i, "Interacción con componente no implementada")
}

// sendResponse envía una respuesta a una interacción
func (h *InteractionHandler) sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		fmt.Println("Error al enviar respuesta:", err)
	}
}

// getOrCreatePlayer obtiene o crea un MusicPlayer para el servidor
func (h *InteractionHandler) getOrCreatePlayer(guildID string) *MusicPlayer {
	h.mu.Lock()
	defer h.mu.Unlock()

	if player, exists := h.players[guildID]; exists {
		return player
	}

	// Crear un nuevo MusicPlayer para el servidor
	player := NewMusicPlayer(
		h.discordMessenger,
		h.messageConsumer,
		h.externalSongService,
		h.songRepo,
		h.songStorage,
		h.stateStorage,
		h.storageAudio,
		h.voiceSessionFactory(guildID), // Sesión de voz específica del servidor
		h.decoderFactory,
	)

	h.players[guildID] = player
	return player
}

// cleanupPlayer cierra y elimina un MusicPlayer cuando ya no es necesario
func (h *InteractionHandler) cleanupPlayer(guildID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if player, exists := h.players[guildID]; exists {
		player.Close() // Método para liberar recursos
		delete(h.players, guildID)
	}
}
