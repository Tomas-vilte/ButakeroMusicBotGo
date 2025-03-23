output "repository_url" {
  description = "URL del repositorio ECR"
  value       = aws_ecr_repository.butakero_bot_prod.repository_url
}