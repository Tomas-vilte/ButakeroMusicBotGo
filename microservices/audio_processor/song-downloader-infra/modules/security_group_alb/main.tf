resource "aws_security_group" "alb" {
  name = "${var.project_name}-${var.environment}-alb-sg"
  description = "Security group para el ALB"
  vpc_id = var.vpc_id

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["${chomp(data.http.myip.response_body)}/32"]
    description = "Permitir el trafico HTTP desde la IP del usuario"
  }

  tags = var.tags
}

data "http" "myip" {
  url = "https://ipv4.icanhazip.com"
}
