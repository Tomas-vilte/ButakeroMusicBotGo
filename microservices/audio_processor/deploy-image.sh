#!/bin/bash

# Aca cambialo por tus credenciales
IMAGE_TAG=${IMAGE_TAG}
REPOSITORY_NAME=${REPOSITORY_NAME}
AWS_REGION=${AWS_REGION}
AWS_ACCOUNT_ID=${AWS_ACCOUNT_ID}
DOCKERFILE_PATH="."

echo "Logueando en AWS ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

echo "Construyendo la imagen Docker para arquitectura ARM64..."
docker buildx create --use
docker buildx build --platform linux/arm64 -t $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$REPOSITORY_NAME:$IMAGE_TAG $DOCKERFILE_PATH --push

echo "Limpiando el builder..."
docker buildx rm

echo "¡Imagen Docker construida y subida a ECR exitosamente!"
