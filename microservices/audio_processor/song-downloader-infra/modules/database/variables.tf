variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "dynamodb_table_operations_tags" {
  description = "Etiqueta para la tabla de operacion"
  type = map(string)
}

variable "dynamodb_table_songs_tag" {
  description = "Etiqueta para la tabla de canciones"
  type = map(string)
}