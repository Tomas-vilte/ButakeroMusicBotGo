variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "retention_in_days" {
  description = "Periodo de retencion de logs en dias"
  type        = number
  default     = 30
}

variable "tags" {
  description = "Etiquetas comunes para el recurso de cloudWatch logs"
  type        = map(string)
  default     = {}
}
