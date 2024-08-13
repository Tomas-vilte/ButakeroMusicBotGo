package observer

import "github.com/bwmarrin/discordgo"

// VoicePresenceNotifier es un tipo que implementa la interfaz Subject. Se encarga de mantener
// una lista de observadores y notificarles sobre los cambios en la presencia de los usuarios
// en los canales de voz. Puede agregar, quitar y notificar a los observadores registrados.
type VoicePresenceNotifier struct {
	observers []VoicePresenceObserver
}

func NewVoicePresenceNotifier() *VoicePresenceNotifier {
	return &VoicePresenceNotifier{}
}

// AddObserver añade un nuevo observador a la lista de observadores del notifier.
// Los observadores registrados recibirán notificaciones sobre los cambios en la presencia
// de los usuarios en los canales de voz.
func (n *VoicePresenceNotifier) AddObserver(o VoicePresenceObserver) {
	n.observers = append(n.observers, o)
}

// RemoveObserver elimina un observador de la lista de observadores del notifier.
// Después de eliminarlo, el observador ya no recibirá notificaciones sobre los cambios
// en la presencia de los usuarios en los canales de voz.
func (n *VoicePresenceNotifier) RemoveObserver(o VoicePresenceObserver) {
	for i, observer := range n.observers {
		if observer == o {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			break
		}
	}
}

// NotifyObservers envía una notificación a todos los observadores registrados sobre
// un cambio en el estado del canal de voz. Cada observador recibe un objeto de tipo
// *discordgo.VoiceStateUpdate con la información actualizada sobre la presencia de los
// usuarios en los canales de voz.
func (n *VoicePresenceNotifier) NotifyObservers(vs *discordgo.VoiceStateUpdate) {
	for _, observer := range n.observers {
		observer.UpdatePresence(vs)
	}
}
