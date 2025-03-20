#!/bin/bash

DOCKER_HUB_USERNAME="tomasvilte"
VERSION="1.0.0"

echo "Introduce tu contraseña de Docker Hub:"
read -s DOCKER_HUB_PASSWORD

echo "Construyendo la imagen para audio_processor..."
cd microservices/audio_processor
docker build -t ${DOCKER_HUB_USERNAME}/audio_processor:${VERSION} -f dockerfile.local .
cd ../..

echo "Construyendo la imagen para bot..."
cd microservices/butakero_bot
docker build -t ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION} -f Dockerfile .

echo "Construyendo la imagen del bot para debug..."
docker build -t ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION}-debug -f Dockerfile.debug .
cd ../..

echo "Verificando las imágenes construidas..."
docker images | grep ${DOCKER_HUB_USERNAME}

echo "Iniciando sesión en Docker Hub..."
echo ${DOCKER_HUB_PASSWORD} | docker login -u ${DOCKER_HUB_USERNAME} --password-stdin

echo "Subiendo audio_processor a Docker Hub..."
docker push ${DOCKER_HUB_USERNAME}/audio_processor:${VERSION}

echo "Subiendo bot a Docker Hub..."
docker push ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION}

echo "Subiendo bot debug a Docker Hub..."
docker push ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION}-debug

echo "¡Proceso completado!"