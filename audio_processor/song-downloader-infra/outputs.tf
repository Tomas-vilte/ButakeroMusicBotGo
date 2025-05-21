output "region" {
  description = "AWS Region"
  value       = var.aws_region
}

output "dynamodb_songs_table" {
  description = "Nombre de la tabla DynamoDB para canciones"
  value       = module.database.songs_table_name
}

output "s3_bucket_name" {
  description = "Nombre del bucket S3"
  value       = module.storage.bucket_name
}

output "secret_arn" {
  description = "ARN del secreto en Secrets Manager"
  value       = module.secret_manager.secret_arn
}

output "ecs_cluster_name" {
  description = "Nombre del cluster ECS"
  value       = module.ecs.cluster_name
}

output "secret_name" {
  description = "Nombre del secreto en Secret Manager"
  value = module.secret_manager.secret_name
}

output "ecs_service_name" {
  description = "Nombre del servicio ECS"
  value       = module.ecs.service_name
}

output "vpc_id" {
  description = "ID del VPC"
  value       = module.networking.vpc_id
}

output "sg_alb_id" {
  value = module.security_groups.security_group_alb_id
  description = "ID del security group del ALB"
}

output "public_subnet_ids" {
  value = module.networking.public_subnet_ids
  description = "Lista de IDs de subnets públicas"
}

# output "private_subnet_ids" {
#   value = module.networking.private_subnet_ids
#     description = "Lista de IDs de subnets privadas"
# }