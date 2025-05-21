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

resource "null_resource" "docker_push" {
  provisioner "local-exec" {
    command = <<-EOT
      cd ..
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com
      docker buildx create --use --name mybuilder-audio-processor
      docker buildx build --platform linux/arm64 --build-arg ENV=aws -t ${aws_ecr_repository.app.repository_url}:${var.image_tag} --push --no-cache ${var.docker_context_path}
      docker buildx rm mybuilder-audio-processor || true
      docker volume ls -q --filter name=buildx_buildkit | xargs -r docker volume rm 2>/dev/null || true
    EOT
  }
  triggers = {
    always_run = timestamp()
  }
}