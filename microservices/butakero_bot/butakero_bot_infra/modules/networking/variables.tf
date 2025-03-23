variable "remote_state_bucket" {
  description = "Nombre del bucket S3 donde se almacena el estado remoto"
  type        = string
}

variable "remote_state_key" {
  description = "Clave del archivo de estado remoto en S3"
  type        = string
}

variable "aws_region" {
  description = "Región de AWS"
  type        = string
  default     = "us-east-1"
}