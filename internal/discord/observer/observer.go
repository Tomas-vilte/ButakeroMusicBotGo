package observer

import "github.com/bwmarrin/discordgo"

type (

	// VoicePresenceObserver es una interfaz para los observadores que quieren recibir actualizaciones
	// sobre el estado de los canales de voz. Cualquier tipo que implemente esta interfaz podrá
	// recibir notificaciones cuando haya cambios en la presencia de los usuarios en los canales de voz.
	VoicePresenceObserver interface {
		// UpdatePresence es el método que será llamado por el sujeto para notificar al observador
		// sobre un cambio en el estado del canal de voz.
		UpdatePresence(voiceState *discordgo.VoiceStateUpdate)
	}

	// Subject es una interfaz para los objetos que mantienen una lista de observadores y les avisan
	// sobre los cambios en su estado. Un sujeto puede agregar, quitar y notificar a los observadores
	// registrados.
	Subject interface {
		// AddObserver agrega un nuevo observador a la lista del sujeto. Los observadores registrados
		// recibirán notificaciones sobre los cambios en el estado del sujeto.
		AddObserver(observer VoicePresenceObserver)

		// RemoveObserver elimina un observador de la lista del sujeto. Después de eliminarlo, el sujeto
		// no notificará más a este observador sobre los cambios.
		RemoveObserver(observer VoicePresenceObserver)

		// NotifyObservers avisa a todos los observadores registrados sobre un cambio en el estado
		// del sujeto. Los observadores recibirán un objeto de tipo *discordgo.VoiceStateUpdate con la
		// información actualizada sobre el estado del canal de voz.
		NotifyObservers(vs *discordgo.VoiceStateUpdate)
	}
)
