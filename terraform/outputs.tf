output "ec2_instance_id" {
  description = "ID de la instancia EC2"
  value       = module.ec2.ec2_instance_id
}

output "ec2_instance_public_ip" {
  description = "Dirección IP pública de la instancia EC2"
  value       = module.ec2.ec2_instance_public_ip
}