variable "vpc_id" {
  description = "ID de la VPC donde se creará la subred"
  type        = string
}

variable "subnet_cidr_block" {
  description = "CIDR block para la subred"
  type        = string
  default     = "10.0.1.0/24"
}

variable "internet_gateway_id" {
  description = "ID de la puerta de enlace de Internet"
  type        = string
}