variable "project_name" {
  description = "Nombre del proyecto"
  type = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type = string
}

variable "tags" {
  description = "Etiquetas"
  type = map(string)
}