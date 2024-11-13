variable "project_name" {
  description = "Nombre del proyecto"
  type = string
  default = "music-downloader"
}

variable "environment" {
  description = "Ambiente de despliegue"
  type = string
}

variable "tags" {
  description = "Etiquetas para los recursos"
  type        = map(string)
  default     = {}
}

variable "vpc_id" {
  description = "VPC ID donde se crea el Security Group"
  type        = string
}
