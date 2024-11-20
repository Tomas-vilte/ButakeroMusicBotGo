output "vpc_id" {
  description = "ID del VPC"
  value = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs de las subnets publicas"
  value = aws_subnet.public[*].id
}