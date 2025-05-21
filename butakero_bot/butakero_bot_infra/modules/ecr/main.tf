resource "aws_ecr_repository" "butakero_bot_prod" {
  name = var.repository_name
  force_delete = true
  image_tag_mutability = "MUTABLE"


  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
  }
}

resource "aws_ecr_lifecycle_policy" "app" {
  repository = aws_ecr_repository.butakero_bot_prod.name

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

resource "null_resource" "build_and_push_image" {
  provisioner "local-exec" {
    command = <<EOT
      cd ..
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com
      docker buildx create --use --name mybuilder-butakero
      docker buildx build --platform linux/arm64 --build-arg ENV=bot_aws -t ${aws_ecr_repository.butakero_bot_prod.repository_url}:${var.image_tag} --push --no-cache ${var.docker_context_path}
      docker buildx rm mybuilder-butakero || true
      docker volume ls -q --filter name=buildx_buildkit | xargs -r docker volume rm 2>/dev/null || true
    EOT
  }

  triggers = {
    always_run = timestamp()
  }
}