variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "tags_sqs_queue" {
  description = "Etiqueta para SQS"
  type = map(string)
}