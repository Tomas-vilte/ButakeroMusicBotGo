variable "secret_name" {
  description = "Nombre del secreto en Secrets Manager"
  type        = string
}

variable "secret_arn" {
  description = "ARN del secreto en Secrets Manager"
  type        = string
}

variable "command_prefix" {
  description = "Prefijo del comando para el bot de Discord"
  type        = string
  sensitive   = true
}

variable "discord_token" {
  description = "Token de autenticación del bot de Discord"
  type        = string
  sensitive   = true
}