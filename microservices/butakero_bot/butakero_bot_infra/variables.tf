variable "aws_region" {
  description = "Región de AWS"
  type        = string
  default     = "us-east-1"
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