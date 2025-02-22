package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/bwmarrin/discordgo"
	"sync"
)

type DiscordPresenceNotifier struct {
	observers []ports.VoicePresenceObserver
	mu        sync.RWMutex
}

func NewDiscordPresenceNotifier() *DiscordPresenceNotifier {
	return &DiscordPresenceNotifier{
		observers: make([]ports.VoicePresenceObserver, 0),
	}
}

func (n *DiscordPresenceNotifier) AddObserver(o ports.VoicePresenceObserver) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.observers = append(n.observers, o)
}

func (n *DiscordPresenceNotifier) RemoveObserver(o ports.VoicePresenceObserver) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, observer := range n.observers {
		if observer == o {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			break
		}
	}
}

func (n *DiscordPresenceNotifier) NotifyAll(vs *discordgo.VoiceStateUpdate) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, observer := range n.observers {
		observer.UpdatePresence(vs)
	}
}
