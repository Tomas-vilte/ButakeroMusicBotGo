package discord

type InteractionResponseType int

const (
	InteractionResponsePong                             InteractionResponseType = 1
	InteractionResponseChannelMessageWithSource         InteractionResponseType = 4
	InteractionResponseDeferredChannelMessageWithSource InteractionResponseType = 5
	InteractionResponseDeferredMessageUpdate            InteractionResponseType = 6
	InteractionResponseUpdateMessage                    InteractionResponseType = 7
	InteractionApplicationCommandAutocompleteResult     InteractionResponseType = 8
	InteractionResponseModal                            InteractionResponseType = 9
)

type (
	Interaction struct {
		ID        string
		AppID     string
		ChannelID string
		GuildID   string
		Member    *Member
		Token     string
	}

	Member struct {
		UserID   string
		Username string
	}

	InteractionResponse struct {
		Type    InteractionResponseType
		Content string
		Embeds  []*Embed
	}

	Embed struct {
		Title       string
		Description string
		Color       int
		Fields      []*EmbedField
		Thumbnail   *EmbedThumbnail
		Footer      *EmbedFooter
	}

	EmbedField struct {
		Name   string
		Value  string
		Inline bool
	}

	EmbedThumbnail struct {
		URL    string
		Width  int
		Height int
	}

	WebhookParams struct {
		Content string
		Embeds  []*Embed
	}

	WebhookEdit struct {
		Content *string
		Embeds  []*Embed
	}

	EmbedFooter struct {
		Text string
	}
)
