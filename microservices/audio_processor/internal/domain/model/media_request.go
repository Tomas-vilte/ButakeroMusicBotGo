package model

import "time"

type MediaRequest struct {
	InteractionID string    `json:"interaction_id"`
	UserID        string    `json:"user_id"`
	Song          string    `json:"song"`
	ProviderType  string    `json:"provider_type"`
	Timestamp     time.Time `json:"timestamp"`
}
