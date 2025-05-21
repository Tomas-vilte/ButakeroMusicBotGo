output "secret_arn" {
  description = "ARN del secreto en Secrets Manager"
  value       = var.secret_arn
}

output "secret_name" {
  value = var.secret_name
  description = "Nombre del secreto en Secret Manager"
}