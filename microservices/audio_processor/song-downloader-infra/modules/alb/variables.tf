variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "vpc_id" {
  description = "ID del VPC"
  type = string
}

variable "subnet_ids" {
  description = "IDs de las subnets"
  type = list(string)
}

variable "container_port" {
  description = "Puerto del contenedor"
  type = number
  default = 8080
}

variable "logs_bucket" {
  description = "Bucket para logs del ALB"
  type = string
}

variable "tags" {
  description = "Tags para recursos"
  type = map(string)
  default = {}
}

variable "security_group_alb" {
  description = "ID del Security Group para el ALB"
  type        = string
}