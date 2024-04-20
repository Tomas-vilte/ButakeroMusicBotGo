package music

type Song struct {
	ChannelID string
	UserName  string
	ID        string
	VideoID   string
	Title     string
	VideoURL  string
}

type PkgSong struct {
	data Song
	v    *VoiceInstance
}

func globalPlay(songSig chan PkgSong) {
	for {
		select {
		case song := <-songSig:
			go song.v.PlayQueue(song.data)
		}
	}
}
