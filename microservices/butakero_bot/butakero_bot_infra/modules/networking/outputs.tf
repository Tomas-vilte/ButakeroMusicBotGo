output "vpc_id" {
  description = "ID de la VPC"
  value       = data.terraform_remote_state.networking.outputs.vpc_id
}

output "public_subnet_ids" {
  description = "Lista de subnets privadas"
  value       = data.terraform_remote_state.networking.outputs.public_subnet_ids
}

output "private_subnet_ids" {
  value = data.terraform_remote_state.networking.outputs.private_subnet_ids
    description = "Lista de subnets privadas"
}