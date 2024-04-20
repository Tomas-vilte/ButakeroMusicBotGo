package music

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"log"
	"sync"
)

type VoiceInstance struct {
	voice      *discordgo.VoiceConnection
	session    *discordgo.Session
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	queueMutex sync.Mutex
	nowPlaying Song
	queue      []Song
	GuildID    string
	speaking   bool
	pause      bool
	stop       bool
	skip       bool
}

func (v *VoiceInstance) Skip() bool {
	if v.speaking {
		if v.pause {
			log.Println("Se ha solicitado saltar la canción, pero la reproducción está pausada.")
			return true
		} else {
			if v.encoder != nil {
				v.encoder.Cleanup()
				log.Println("Se ha limpiado el encoder al saltar la canción.")
			}
		}
	}
	return false
}

func (v *VoiceInstance) Stop() {
	v.stop = true
	if v.encoder != nil {
		v.encoder.Cleanup()
		log.Println("Se ha detenido la reproducción y limpiado el encoder.")
	}
}

func (v *VoiceInstance) QueueAdd(song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = append(v.queue, song)
	log.Printf("Se ha añadido la canción '%s' a la cola de reproducción.", song.Title)
}

func (v *VoiceInstance) QueueGetSong() (song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		log.Printf("Se ha obtenido la próxima canción de la cola de reproducción: '%s'.", song.Title)
		return v.queue[0]
	}
	return
}

func (v *VoiceInstance) QueueRemoveFirst() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		removedSong := v.queue[0]
		v.queue = v.queue[1:]
		log.Printf("Se ha eliminado la canción '%s' de la cola de reproducción.", removedSong.Title)
	}
}

func (v *VoiceInstance) QueueClear() {
	v.queueMutex.Unlock()
	defer v.queueMutex.Lock()
	v.queue = []Song{}
	clearedQueue := len(v.queue)
	log.Printf("Se ha vaciado la cola de reproducción. Se eliminaron %d canciones.", clearedQueue)
}

func (v *VoiceInstance) PlayQueue(song Song, client discord.DiscordClient) {
	v.QueueAdd(song)

	if v.speaking {
		return
	}

	go func() {
		for {
			if len(v.queue) == 0 {
				// La cola de canciones está vacía
				log.Println("La cola esta vacia")
				client.SendChannelMessage(v.nowPlaying.ChannelID, "[musica]fin de la cola")
				return
			}

			v.nowPlaying = v.QueueGetSong()
			client.SendChannelMessage(v.nowPlaying.ChannelID, "[musica] escuchando:"+v.nowPlaying.Title)
			v.stop = false
			v.skip = false
			v.speaking = true
			v.pause = false
			err := v.voice.Speaking(true)
			if err != nil {
				log.Fatalf("No se pudo enviar la notificación de conversación: %v", err)
				return
			}

			v.DCA(v.nowPlaying.VideoURL)

			v.QueueRemoveFirst()
			if v.stop {
				v.QueueClear()
			}
			v.stop = false
			v.skip = false
			v.speaking = false

			err = v.voice.Speaking(false)
			if err != nil {
				log.Fatalf("Failed to stop sending speaking notification: %v", err)
			}
		}
	}()
}

func (v *VoiceInstance) DCA(url string) {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 64
	opts.Application = "lowdelay"

	encondeSession, err := dca.EncodeFile(url, opts)
	if err != nil {
		log.Printf("Error al codificar el archivo: %v", err)
		return
	}

	v.encoder = encondeSession
	done := make(chan error)
	v.stream = dca.NewStream(encondeSession, v.voice, done)
	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Printf("Error al reproducir el archivo codificado: %v", err)
			}

			encondeSession.Cleanup()
			return
		}
	}
}
