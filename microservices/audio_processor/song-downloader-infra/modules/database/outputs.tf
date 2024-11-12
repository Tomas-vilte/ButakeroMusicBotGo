output "songs_table_name" {
  description = "Nombre de la tabla de canciones"
  value = aws_dynamodb_table.songs.name
}

output "operations_table_name" {
  description = "Nombre de la tabla de operaciones"
  value = aws_dynamodb_table.operations.name
}

output "table_arns" {
  description = "ARNs de las tablas DynamoDB"
  value = [
    aws_dynamodb_table.songs.arn,
    aws_dynamodb_table.operations.arn
  ]
}