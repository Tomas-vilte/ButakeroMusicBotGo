output "ec2_instance_id" {
  description = "ID de la instancia EC2"
  value       = aws_instance.instancia_ec2.id
}

output "ec2_instance_public_ip" {
  description = "Dirección IP pública de la instancia EC2"
  value       = aws_instance.instancia_ec2.public_ip
}