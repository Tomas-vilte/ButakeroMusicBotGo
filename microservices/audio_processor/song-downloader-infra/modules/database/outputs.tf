output "songs_table_name" {
  description = "Nombre de la tabla de canciones"
  value = aws_dynamodb_table.songs.name
}

output "table_arns" {
  description = "ARNs de las tablas DynamoDB"
  value = [
    aws_dynamodb_table.songs.arn,
  ]
}