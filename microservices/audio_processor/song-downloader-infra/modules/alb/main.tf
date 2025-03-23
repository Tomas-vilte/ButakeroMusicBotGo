resource "aws_lb" "main" {
  name = "${var.project_name}-alb-${var.environment}"
  internal = true
  load_balancer_type = "application"
  security_groups = [var.security_group_alb]
  subnets = var.private_subnet_ids

  enable_deletion_protection = false
  preserve_host_header = true

  access_logs {
    bucket = var.logs_bucket
    prefix = "alb/alb-prod"
    enabled = true
  }

  tags = var.tags
}

resource "aws_lb_target_group" "main" {
  name = "${var.project_name}-tg-${var.environment}"
  port = var.container_port
  protocol = "HTTP"
  vpc_id = var.vpc_id
  target_type = "ip"
  deregistration_delay = 30

  health_check {
    enabled = true
    healthy_threshold = 3
    interval = 30
    matcher = "200"
    path = "/api/v1/health"
    port = "traffic-port"
    protocol = "HTTP"
    timeout = 5
    unhealthy_threshold = 3
  }

  tags = var.tags
}

resource "aws_lb_listener" "main" {
  load_balancer_arn = aws_lb.main.arn
  port = "80"
  protocol = "HTTP"

  default_action {
    type = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}