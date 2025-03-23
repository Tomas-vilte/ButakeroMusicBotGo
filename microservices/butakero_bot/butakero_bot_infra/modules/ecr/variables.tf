variable "repository_name" {
  description = "Nombre del repositorio ECR"
  type        = string
  default     = "butakero-bot-prod"
}

variable "aws_region" {
  description = "Región de AWS"
  type        = string
  default     = "us-east-1"
}

variable "aws_account_id" {
  description = "ID de la cuenta de AWS"
  type        = string
}

variable "image_tag" {
  description = "Tag de la imagen Docker"
  type        = string
  default     = "latest"
}

variable "docker_context_path" {
  description = "Ruta del contexto de Docker para construir la imagen"
  type        = string
  default     = "."
}