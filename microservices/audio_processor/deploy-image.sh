#!/bin/bash

# Aca cambialo por tus credenciales
REPOSITORY_URL=${REPOSITORY_URL}
REPOSITORY_NAME="butakero-music-download-prod"
AWS_REGION=${AWS_REGION}
DOCKERFILE_PATH="."

echo "Logueando en AWS ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $REPOSITORY_URL

echo "Construyendo la imagen Docker para arquitectura ARM64..."
docker buildx create --use
docker buildx build --platform linux/arm64 -t $REPOSITORY_URL/$REPOSITORY_NAME:latest --push .

echo "Limpiando el builder..."
docker buildx rm

echo "¡Imagen Docker construida y subida a ECR exitosamente!"
