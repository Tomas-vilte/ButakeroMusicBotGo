resource "aws_ecr_repository" "butakero_bot_prod" {
  name = var.repository_name
}

resource "null_resource" "build_and_push_image" {
  provisioner "local-exec" {
    command = <<EOT
      cd ..
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com
      docker buildx create --use
      docker buildx build --platform linux/arm64 --build-arg ENV=bot_aws -t ${aws_ecr_repository.butakero_bot_prod.repository_url}:${var.image_tag} --push ${var.docker_context_path}
    EOT
  }

  triggers = {
    always_run = timestamp()
  }
}