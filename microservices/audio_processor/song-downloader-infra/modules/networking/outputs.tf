output "vpc_id" {
  description = "ID del VPC"
  value = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs de las subnets publicas"
  value = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs de las subnets privadas"
  value = aws_subnet.private[*].id
}