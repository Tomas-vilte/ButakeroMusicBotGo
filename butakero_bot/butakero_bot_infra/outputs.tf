output "ecs_service_name" {
  description = "Nombre del servicio ECS"
  value = module.ecs.ecs_service_name
}

output "ecr_repository_url" {
  description = "URL del repositorio ECR"
  value       = module.ecr.repository_url
}

output "secret_arn" {
  description = "ARN del secreto en Secrets Manager"
  value       = module.secret_manager.secret_arn
}

output "secret_name" {
  value = module.secret_manager.secret_name
    description = "Nombre del secreto en Secret Manager"
}

