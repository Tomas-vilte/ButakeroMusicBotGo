output "vpc_id" {
  description = "ID de la VPC"
  value       = aws_vpc.vpc.id
}

output "internet_gateway_id" {
  description = "ID de la puerta de enlace de Internet"
  value       = aws_internet_gateway.internet_gateway.id
}