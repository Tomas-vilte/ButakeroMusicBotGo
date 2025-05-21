package entity

type HealthStatus string

const (
	StatusOperational HealthStatus = "operational"
	StatusDegraded    HealthStatus = "degraded"
	StatusDown        HealthStatus = "down"
)

type (
	HealthResponse struct {
		Status    HealthStatus   `json:"status"`
		Discord   DiscordHealth  `json:"discord"`
		ServiceB  ServiceBHealth `json:"service_b"`
		Timestamp string         `json:"timestamp"`
		Message   string         `json:"message,omitempty"`
		Version   string         `json:"version,omitempty"`
	}

	DiscordHealth struct {
		Connected          bool    `json:"connected"`
		HeartbeatLatencyMS float64 `json:"heartbeat_latency_ms"`
		Guilds             int     `json:"guilds"`
		VoiceConnections   int     `json:"voice_connections"`
		SessionID          string  `json:"session_id,omitempty"`
		CheckDurationMS    float64 `json:"check_duration_ms,omitempty"`
		Error              string  `json:"error,omitempty"`
	}

	ServiceBHealth struct {
		Connected bool   `json:"connected"`
		LatencyMS int    `json:"latency_ms"`
		Status    string `json:"status,omitempty"`
		Error     string `json:"error,omitempty"`
	}
)

func DetermineOverallStatus(discord DiscordHealth, serviceB ServiceBHealth) (HealthStatus, string) {
	if !discord.Connected && !serviceB.Connected {
		return StatusDown, "Servicios críticos no disponibles"
	}

	if !discord.Connected {
		return StatusDegraded, "Discord no está conectado"
	}

	if !serviceB.Connected {
		return StatusDegraded, "Service B no está disponible"
	}

	if discord.HeartbeatLatencyMS > 1000 {
		return StatusDegraded, "Discord presenta alta latencia"
	}

	if serviceB.LatencyMS > 1000 {
		return StatusDegraded, "Service B presenta alta latencia"
	}

	return StatusOperational, ""
}
