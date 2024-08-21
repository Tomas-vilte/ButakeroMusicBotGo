package types

type Metadata struct {
	Opus   *OpusMetadata   `json:"opus"`
	Origin *OriginMetadata `json:"origin"`
}

type OriginMetadata struct {
	Source   string `json:"source"`
	Bitrate  int    `json:"abr"`
	Channels int    `json:"channels"`
	Encoding string `json:"encoding"`
	Url      string `json:"url"`
}

type OpusMetadata struct {
	Bitrate     int    `json:"abr"`
	SampleRate  int    `json:"sample_rate"`
	Application string `json:"mode"`
	FrameSize   int    `json:"frame_size"`
	Channels    int    `json:"channels"`
	VBR         bool   `json:"vbr"`
}
