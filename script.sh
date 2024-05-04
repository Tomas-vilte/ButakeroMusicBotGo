#!/bin/bash

# Crear un directorio para los binarios
mkdir -p dependencies_layer/bin

# Descargar ffmpeg
echo "Descargando ffmpeg..."
wget -O dependencies_layer/bin/ffmpeg https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz
tar -xvf dependencies_layer/bin/ffmpeg -C dependencies_layer/bin --strip-components=1 --wildcards '*/ffmpeg' '*/ffprobe'
rm dependencies_layer/bin/ffmpeg-release-amd64-static.tar.xz


# Descargar yt-dlp
echo "Descargando yt-dlp..."
wget -O dependencies_layer/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
chmod +x dependencies_layer/bin/yt-dlp

# Comprimir los binarios en un archivo zip
cd dependencies_layer
zip -r ffmpeg_yt_dlp_layer.zip bin
