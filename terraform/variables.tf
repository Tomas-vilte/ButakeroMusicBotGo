variable "vpc_cidr_block" {
  description = "CIDR block para la VPC"
  type        = string
}

variable "subnet_cidr_block" {
  description = "CIDR block para la subred"
  type        = string
}

variable "ami_id" {
  description = "ID de la AMI de Amazon Linux 2"
  type        = string
}

variable "instance_type" {
  description = "Tipo de instancia EC2"
  type        = string
}

variable "key_name" {
  description = "Nombre de la llave SSH"
  type        = string
}

variable "availability_zone" {
  description = "Zona en donde se va a crear la instancia"
  type        = string
}
