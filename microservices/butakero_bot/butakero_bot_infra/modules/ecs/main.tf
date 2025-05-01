resource "aws_ecs_task_definition" "music_bot" {
  family                   = "music-bot"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([
    {
      name      = "music-bot"
      image     = var.music_bot_image
      essential = true
      portMappings = [
        {
          containerPort = var.container_port
          protocol = "tcp"
        }
      ]
      environment = [
        {
          name  = "AWS_REGION"
          value = var.aws_region
        },
        {
          name  = "AWS_SECRET_NAME"
          value = var.aws_secret_name
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-create-group  = "true",
          awslogs-group         = aws_cloudwatch_log_group.music_bot.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "music-bot"
        }
      }
      healthCheck = {
        command =   ["CMD-SHELL", "curl -f http://localhost:${var.container_port}/api/v1/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }

      essential = true
    }
  ])
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  tags = {
    "Project" = "music-bot"
  }
}

resource "aws_ecs_service" "music_bot_service" {
  name            = "music-bot-service"
  cluster         = var.cluster_name
  task_definition = aws_ecs_task_definition.music_bot.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"
  force_new_deployment = true
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent = 200
  health_check_grace_period_seconds = 80
  enable_execute_command = true

  deployment_circuit_breaker {
    enable = true
    rollback = true
  }

  network_configuration {
    subnets          = var.public_subnet_ids
    security_groups  = [var.security_group_id]
    assign_public_ip = true
  }

  depends_on = [aws_ecs_task_definition.music_bot]
}

resource "aws_cloudwatch_log_group" "music_bot" {
  name = "/ecs/music-bot"
}