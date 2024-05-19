variable "ami_id" {
  description = "ID de la AMI de Amazon Linux 2"
  type        = string
  default     = "ami-0eac975a54dfee8cb"
}

variable "instance_type" {
  description = "Tipo de instancia EC2"
  type        = string
  default     = "t4g.small"
}

variable "key_name" {
  description = "Nombre de la llave SSH"
  type        = string
  default     = "llave-ssh"
}

variable "subnet_id" {
  description = "ID de la subred donde se creará la instancia EC2"
  type        = string
}

variable "security_group_id" {
  description = "ID del grupo de seguridad para la instancia EC2"
  type        = string
}
variable "availability_zone" {
  description = "Zona en donde se va crear la instancia"
  type        = string
}
