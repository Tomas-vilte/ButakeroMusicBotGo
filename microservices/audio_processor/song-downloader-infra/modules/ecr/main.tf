resource "aws_ecr_repository" "app" {
  name = "${var.project_name}-${var.environment}"
  image_tag_mutability = "MUTABLE"

  force_delete = true

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
  }

  tags = var.tags
}

resource "aws_ecr_lifecycle_policy" "app" {
  repository = aws_ecr_repository.app.name

  policy = jsonencode({
    rules = [{
        rulePriority = 1
        description = "Mantener las ultimas 30 imagenes"
        selection = {
            tagStatus = "any"
            countType = "imageCountMoreThan"
            countNumber = 30
        }
        action = {
            type = "expire"
        }
    }]
  })
}