package music

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
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
			return true
		} else {
			if v.encoder != nil {
				v.encoder.Cleanup()
			}
		}
	}
	return false
}

func (v *VoiceInstance) Stop() {
	v.stop = true
	if v.encoder != nil {
		v.encoder.Cleanup()
	}
}

func (v *VoiceInstance) QueueAdd(song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = append(v.queue, song)
}

func (v *VoiceInstance) QueueGetSong() (song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		return v.queue[0]
	}
	return
}

func (v *VoiceInstance) QueueRemoveFirst() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		v.queue = v.queue[1:]
	}
}

func (v *VoiceInstance) QueueClear() {
	v.queueMutex.Unlock()
	defer v.queueMutex.Lock()
	v.queue = []Song{}
}

func (v *VoiceInstance) PlayQueue(song Song) {
	v.QueueAdd(song)

	if v.speaking {
		return
	}

	go func() {
		for {
			if len(v.queue) == 0 {
				log.Println("La cola esta vacia")
				discord.Session.ChannelMessageSend(v.nowPlaying.ChannelID, "[musica] fin de la cola")
				// La cola de canciones está vacía
				return
			}

			v.nowPlaying = v.QueueGetSong()
			go discord.Session.ChannelMessageSend(v.nowPlaying.ChannelID, "[musica] escuchando:"+v.nowPlaying.Title)
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

}
