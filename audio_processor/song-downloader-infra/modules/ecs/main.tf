resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-${var.environment}"

  setting {
    name = "containerInsights"
    value = "enabled"
  }

  tags = var.tags
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name
  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base = 1
    weight = 100
    capacity_provider = "FARGATE"
  }
}

resource "aws_ecs_task_definition" "main" {
    family = "${var.project_name}-${var.environment}"
    network_mode = "awsvpc"
    requires_compatibilities = ["FARGATE"]
    cpu = var.task_cpu
    memory = var.task_memory
    execution_role_arn = var.execution_role_arn
    task_role_arn = var.task_role_arn

    container_definitions = jsonencode([
        {
            name = "${var.project_name}-container"
            image = "${var.ecr_repository_url}:latest"

            environment = [
              {
                name = "AWS_REGION"
                value = var.aws_region
              },
              {
                name = "AWS_SECRET_NAME"
                value = var.secret_name
              },
              {
                name = "ENVIRONMENT"
                value = var.environment
              },
            ]

            logConfiguration = {
                logDriver = "awslogs"
                options = {
                    "awslogs-group" = var.cloudwatch_log_group
                    "awslogs-create-group": "true"
                    "awslogs-region" = var.aws_region
                    "awslogs-stream-prefix" = "ecs"
                }
            }

            portMappings = [
                {
                    containerPort = var.container_port
                    protocol = "tcp"
                }
            ]
            
            healthCheck = {
                command = ["CMD-SHELL", "curl -f http://localhost:${var.container_port}/api/v1/health || exit 1"]
                interval = 30
                timeout = 5
                retries = 3
                startPeriod = 60
            }

            essential = true
        }
    ])

    runtime_platform {
      operating_system_family = "LINUX"
      cpu_architecture = "ARM64"
    }

    tags = var.tags
}

resource "aws_ecs_service" "main" {
  name = "${var.project_name}-service-${var.environment}"
  cluster = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count = var.service_desired_count
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent = 200
  launch_type = "FARGATE"
  scheduling_strategy = "REPLICA"
  platform_version = "LATEST"
  health_check_grace_period_seconds = 80
  enable_execute_command = true

  network_configuration {
    security_groups = [var.ecs_security_group_id]
    subnets = var.subnet_ids
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name = "${var.project_name}-container"
    container_port = var.container_port
  }

  deployment_circuit_breaker {
    enable = true
    rollback = true
  }

  tags = var.tags
}

resource "aws_appautoscaling_target" "ecs" {
  max_capacity = var.max_capacity
  min_capacity = var.min_capacity
  resource_id = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace = "ecs"
}

resource "aws_appautoscaling_policy" "cpu" {
  name = "${var.project_name}-cpu-autoscaling"
  policy_type = "TargetTrackingScaling"
  resource_id = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace = aws_appautoscaling_target.ecs.service_namespace
  
  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = var.cpu_threshold
  }
}

resource "aws_appautoscaling_policy" "memory" {
  name = "${var.project_name}-memory-autoscaling"
  policy_type = "TargetTrackingScaling"
  resource_id = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value = var.memory_threshold
  }
}