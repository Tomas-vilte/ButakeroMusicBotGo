variable "ami_id" {
  description = "ID de la AMI de Amazon Linux 2"
  type        = string
  default     = "ami-0e001c9271cf7f3b9"
}

variable "instance_type" {
  description = "Tipo de instancia EC2"
  type        = string
  default     = "t3.micro"
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