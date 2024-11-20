output "region" {
  description = "AWS Region"
  value       = var.aws_region
}

output "dynamodb_songs_table" {
  description = "Nombre de la tabla DynamoDB para canciones"
  value       = module.database.songs_table_name
}

output "dynamodb_operations_table" {
  description = "Nombre de la tabla DynamoDB para operaciones"
  value       = module.database.operations_table_name
}

output "s3_bucket_name" {
  description = "Nombre del bucket S3"
  value       = aws_s3_bucket.storage.id
}

output "dns_alb" {
  description = "DNS Del ALB"
  value = module.alb.alb_dns_name
}