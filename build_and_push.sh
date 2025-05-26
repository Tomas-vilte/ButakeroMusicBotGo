#!/bin/bash

set -e

DOCKER_HUB_USERNAME="tomasvilte"
VERSION_BOT="1.2.0"
VERSION_AUDIO_SERVICE="1.1.0"
PLATFORMS="linux/amd64,linux/arm64"

echo "Introduce tu contraseña de Docker Hub:"
read -s DOCKER_HUB_PASSWORD

echo "Configurando Docker buildx para construcción multi-arquitectura..."
if ! docker buildx ls | grep -q "multiarch-builder"; then
    echo "Creando nuevo builder multi-arquitectura..."
    docker buildx create --name multiarch-builder --driver docker-container --bootstrap
fi

docker buildx use multiarch-builder

echo "Iniciando sesión en Docker Hub..."
if ! echo "${DOCKER_HUB_PASSWORD}" | docker login -u "${DOCKER_HUB_USERNAME}" --password-stdin; then
    echo "Error al iniciar sesión en Docker Hub"
    exit 1
fi

echo "Construyendo y subiendo la imagen para audio_processor (${PLATFORMS})..."
cd audio_processor/
docker buildx build \
    --platform "${PLATFORMS}" \
    --build-arg ENV=local \
    -t "${DOCKER_HUB_USERNAME}/audio_processor:${VERSION_AUDIO_SERVICE}" \
    -f Dockerfile \
    --push \
    .

echo "Construyendo y subiendo la imagen para bot (${PLATFORMS})..."
cd ../
cd butakero_bot/
docker buildx build \
    --platform "${PLATFORMS}" \
    --build-arg ENV=bot_local \
    -t "${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT}" \
    -f Dockerfile \
    --push \
    .

echo "Construyendo y subiendo la imagen del bot para debug (${PLATFORMS})..."
docker buildx build \
    --platform "${PLATFORMS}" \
    --build-arg ENV=bot_local \
    -t "${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT}-debug" \
    -f Dockerfile.debug \
    --push \
    .

echo "Verificando las imágenes subidas..."
echo "Las siguientes imágenes fueron construidas y subidas para ${PLATFORMS}:"
echo "- ${DOCKER_HUB_USERNAME}/audio_processor:${VERSION_AUDIO_SERVICE}"
echo "- ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT}"
echo "- ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT}-debug"

echo "¡Proceso completado!"