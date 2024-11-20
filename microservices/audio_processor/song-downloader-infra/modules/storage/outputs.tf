output "bucket_name" {
  description = "Nombre del bucket de S3"
  value = aws_s3_bucket.storage.id
}

output "bucket_arn" {
  description = "ARN del bucket S3"
  value       = aws_s3_bucket.storage.arn
}
