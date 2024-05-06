output "security_group_id" {
  description = "ID del grupo de seguridad"
  value       = aws_security_group.security_group.id
}