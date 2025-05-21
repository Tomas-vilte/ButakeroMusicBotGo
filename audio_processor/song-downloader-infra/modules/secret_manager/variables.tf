variable "project_name" {
  description = "Nombre del proyecto"
  type = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type = string
}

variable "tags" {
  type = map(string)
  default = {}
}

variable "secret_name" {
  description = "Nombre del secreto"
  type = string
}

variable "secret_values" {
  type = map(string)
  description = "Secretos"
  default = {}
}