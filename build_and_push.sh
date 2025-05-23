#!/bin/bash

set -e

DOCKER_HUB_USERNAME="tomasvilte"
VERSION_BOT="1.2.0"
VERSION_AUDIO_SERVICE="1.1.0"

echo "Introduce tu contraseña de Docker Hub:"
read -s DOCKER_HUB_PASSWORD

echo "Construyendo la imagen para audio_processor..."
cd audio_processor
docker build -t ${DOCKER_HUB_USERNAME}/audio_processor:${VERSION_AUDIO_SERVICE} --build-arg ENV=local -f Dockerfile .
cd ../..

echo "Construyendo la imagen para bot..."
cd butakero_bot
docker build -t ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT} --build-arg ENV=bot_local -f Dockerfile .

echo "Construyendo la imagen del bot para debug..."
docker build -t ${DOCKER_HUB_USERNAME}/butakero_bot:${VERSION_BOT}-debug --build-arg ENV=bot_local -f Dockerfile.debug .
cd ../..

echo "Verificando las imágenes construidas..."
docker images | grep ${DOCKER_HUB_USERNAME}

echo "Iniciando sesión en Docker Hub..."
if ! echo "${DOCKER_HUB_PASSWORD}" | docker login -u "${DOCKER_HUB_USERNAME}" --password-stdin; then
    echo "Error al iniciar sesión en Docker Hub"
    exit 1
fi

for image in \
    "audio_processor:${VERSION_AUDIO_SERVICE}" \
    "butakero_bot:${VERSION_BOT}" \
    "butakero_bot:${VERSION_BOT}-debug"
do
    echo "Subiendo ${DOCKER_HUB_USERNAME}/${image} a Docker Hub..."
    if ! docker push "${DOCKER_HUB_USERNAME}/${image}"; then
        echo "Error al subir ${image}"
        exit 1
    fi
done

echo "¡Proceso completado!"