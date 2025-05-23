# ButakeroMusicBot

**ButakeroMusicBot** es una solución que armamos con un par de microservicios, pensados para descargar, procesar y subir audios de videos de YouTube a Amazon S3. La idea es que uses tecnologías modernas y que todo esté listo para que lo despliegues donde quieras: en la nube, con Kubernetes o acá nomás, en tu máquina con Docker Compose.

## Tabla de Contenidos 📋

- [Características Principales](#características-principales)
- [Requisitos Previos](#requisitos-previos)
- [Formas de Despliegue](#formas-de-despliegue)
  - [1. Docker Compose](#1-docker-compose)
  - [2. Kubernetes](#2-kubernetes)
  - [3. AWS ECS con Terraform](#3-aws-ecs-con-terraform)
- [Endpoints del API](#endpoints-del-api)
- [Pruebas](#pruebas)
- [Explicación de diagramas de Arquitectura](#explicación-de-diagramas-de-arquitectura)
  - [1. Arquitectura de la Aplicación en AWS ECS](#arquitectura-de-la-aplicación-en-aws-ecs)
  - [2. Arquitectura en Kubernetes](#arquitectura-en-kubernetes)
---

## Características Principales

- 🔍 Búsqueda de videos en YouTube por nombre o URL.
- 🎧 Descarga y procesamiento de audio.
- ☁️ Subida de archivos de audio a Amazon S3.
- 🗃️ Registro de operaciones y metadatos en DynamoDB.
- 🔄 Sistema de reintentos en caso de fallos.
- 📊 Monitoreo y métricas con Prometheus/Grafana
- 🔒 Seguridad y autenticación integrada

## Requisitos Previos

Para arrancar, vas a necesitar esto:

### Requisitos Generales

- Go 1.21 o superior
- Docker y Docker Compose
- Cuenta de AWS con acceso programático
- API Key de YouTube
- Git
---

### Para Despliegue en AWS

- AWS CLI configurado
- Terraform >= 1.0
- Cuenta de AWS con permisos para:
  - ECS
  - ECR
  - S3
  - DynamoDB
  - IAM
  - VPC
  - CloudWatch

---

### Para Despliegue en Kubernetes

- Kubernetes Cluster (1.24+)
- kubectl configurado
- Helm v3
- Minikube (para desarrollo local)

---

## Formas de Despliegue

Te cuento cómo lo podés levantar a nuestro ButakeroMusicBot, que como te decía, ahora son dos microservicios.

### 1. Docker Compose

El despliegue con Docker Compose levanta los siguientes servicios:

- **Butakero-Bot-Service**: El que se encarga de recibir las solicitudes de los usuarios y coordinar.
- **Audio-Processing-Service**: El que hace el laburo pesado de descargar y procesar los audios.
- Zookeeper
- Kafka
- MongoDB

### Configuración Inicial

1. Asegurate de tener Docker y Docker Compose instalados en tu máquina

2. Creá un archivo llamado .env en el mismo directorio donde está el docker-compose.yml. Este archivo tiene que tener estas variables necesarias para que todo ande. ¡Solo necesitamos estas tres, las demás ya están configuradas en los servicios!
```env
YOUTUBE_API_KEY=TU_API_KEY_AQUI
DISCORD_TOKEN=TU_DISCORD_TOKEN_AQUI
COMMAND_PREFIX=bot_music # El prefijo que el bot va a escuchar en Discord (ej: !play)
```

#### Obtener YouTube API Key

Para obtener tu YouTube API Key, sigue estos pasos:

1. Ve a la [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto o selecciona uno existente
3. Habilita la YouTube Data API v3 para tu proyecto
4. Ve a "Credenciales"
5. Crea una nueva credencial de tipo "API Key"
6. Copia la API Key generada y colócala en la variable `YOUTUBE_API_KEY` del archivo `.env`

3. Script de cookies de YouTube (opcional): Si encontras problemas con YouTube, como el error "Sign in to confirm your age", podes generar un archivo de cookies para evitar el bloqueo. Para ello, tenes que ejecutar previamente el script [filter_youtube_cookies.sh](../audio_processor/filter_youtube_cookies.sh) que generará el archivo `yt-cookies.txt`.

```bash
bash filter_youtube_cookies.sh
```

#### Ejecutar con Docker Compose
El proyecto incluye un [docker-compose.yaml](../docker-compose.yaml) que levanta todos los servicios necesarios:
- Zookeeper
- Kafka
- MongoDB

Para iniciar todos los servicios:
```bash
docker compose up -d
```

Para ver los logs de los servicios:
```bash
# Ver todos los logs
docker compose logs -f butakero_bot
docker compose logs -f audio_processor
```

Para detener los servicios:
```bash
docker compose down
```

### 2. Kubernetes

#### Prerrequisitos
- Kubernetes Cluster (1.24+)
- kubectl configurado
- Helm v3
- Minikube (para desarrollo local)

#### Configuración de Credenciales
Antes de ejecutar el despliegue, debes configurar las credenciales de los diferentes baceknd el de backend-secret [secret.yaml](/k8s/bases/backend/download-service/secret.yaml) y discord-bot-secrets [secret.yaml](/k8s/bases//backend/discord-bot/secret.yaml):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-secret
  namespace: backend
type: Opaque
stringData:
  YOUTUBE_API_KEY: "${YOUTUBE_API_KEY}"  # Tu API key de YouTube
  MONGO_USER: "admin-user"     
  MONGO_PASSWORD: "root"
  YT_COOKIES: "${YT_COOKIES}"   # Cookies de YouTube (opcional)

---

apiVersion: v1
kind: Secret
metadata:
  name: discord-bot-secrets
  namespace: backend
type: Opaque
stringData:
  DISCORD_TOKEN: "" # Aca pone tu token de discord

apiVersion: v1
kind: ConfigMap
metadata:
  name: discord-bot-config
  namespace: backend
data:
  COMMAND_PREFIX: "test" # Aca cambialo por el que vos quieras por ej /bot
```

#### Obtener Cookies de YouTube (Opcional)
Para evitar errores de autorización en YouTube ("Sign in to confirm your age"), puedes configurar cookies:
1. Ejecuta el script:
```bash
bash filter_youtube_cookies.sh
```

2. Copia el contenido generado al campo YT_COOKIES del secret.

#### Despliegue Automatizado

El proyecto incluye un script [deploy.sh](../deploy.sh) que automatiza todo el proceso de despliegue. Este script:

1. Crea los namespaces necesarios:
   - Kafka
   - MongoDB
   - Prometheus
   - Backend
   - Otros servicios

2. Configura RBAC y permisos para:
   - Prometheus
   - MongoDB

3. Instala y configura componentes esenciales:
   - Cert-Manager para gestión de certificados SSL/TLS
   - Operador de MongoDB
   - Kafka con Strimzi
   - Prometheus & Grafana para monitorización
   - Backend de la aplicación

Para ejecutar el despliegue completo, simplemente ejecuta:

```bash
bash deploy.sh
```

El script se encargará de:
- Verificar los prerrequisitos
- Crear todos los recursos necesarios
- Esperar a que los pods estén listos
- Verificar el estado final del despliegue

#### Acceso Local a los Endpoints

Una vez completado el despliegue, puedes acceder a los endpoints localmente:

```bash
kubectl port-forward svc/audio-processing-service 8080 -n backend
```

Esto te permitirá acceder al servicio en http://localhost:8080/api/v1/health

#### Verificar el Estado del Despliegue

Para verificar el estado de los pods:

```bash
# Ver pods en el namespace backend
kubectl get pods -n backend

# Ver pods en todos los namespaces
kubectl get pods --all-namespaces
```

#### Logs y Monitoreo

Para ver los logs del backend:

```bash
# Obtener el nombre del pod
kubectl get pods -n backend

# Ver logs
kubectl logs -f butakero-bot-application-hash -n butakero-bot # Reemplazá 'hash' por el nombre de tu pod

kubectl logs -f audio-processing-application-hash -n audio-processing # Reemplazá 'hash' por el nombre de tu pod
```

Para acceder a Grafana:

```bash
kubectl port-forward svc/grafana 3000 -n monitoring
```

### 3. AWS ECS con Terraform

#### Prerrequisitos
- Terraform >= 1.0
- AWS CLI configurado
- Cuenta de AWS con permisos necesarios

#### Configuración del Backend Remoto

> **Advertencia:**  
> Antes de seguir, es **obligatorio** generar las cookies con el script [filter_youtube_cookies.sh](../audio_processor/filter_youtube_cookies.sh). Muchas IPs de las instancias de AWS están **baneadas o bloqueadas** por YouTube, así que vas a tener que usar cookies válidas para evitar problemas de conexión. Si vas a desplegarlo en ECS, asegurate de hacer este paso antes.

1. Creá el bucket S3 (si no lo tenés). Este bucket lo vamos a usar para guardar los estados de Terraform de ambos servicios:
```bash
aws s3api create-bucket \
    --bucket song-download-tf-state \
    --region us-east-1
```

2. Habilitar el versionado del bucket:
```bash
aws s3api put-bucket-versioning \
    --bucket song-download-tf-state \
    --versioning-configuration Status=Enabled
```

3. Creá dos tablas DynamoDB para bloquear el estado de cada despliegue de Terraform. Esto es crucial para que no haya conflictos si varias personas (o procesos) intentan aplicar cambios al mismo tiempo.

* Para el audio_processor (donde vamos a guardar su estado de Terraform):
```bash
aws dynamodb create-table \
    --table-name terraform-lock-table_audio_processing \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region us-east-1
```
* Para el butakero_bot (donde vamos a guardar su estado de Terraform):
```bash
aws dynamodb create-table \
    --table-name terraform_lock_table_music_bot \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region us-east-1
```

#### Configuración del Provider y Backend por Servicio

Te vamos a mostrar la configuración de providers.tf y terraform.example.tfvars para cada uno de nuestros microservicios, ya que los desplegamos por separado.

El archivo [providers.tf](../audio_processor/song-downloader-infra/providers.tf) debe tener esto (Backend para audio_processor):

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket         = "song-download-tf-state"
    key            = "prod/terraform.tfstate" # Clave específica para el audio_processor
    region         = "us-east-1"
    dynamodb_table = "terraform-lock-table_audio_processing" # Tabla específica para el audio_processor
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}
```

Y el [terraform.example.tfvars](../audio_processor/song-downloader-infra/terraform.example.tfvars) de audio_processor:

```hcl
# Configuración básica
aws_region = "us-east-1"
project_name = "butakero-music-download-audio-processor" # Nombre de proyecto ajustado
environment = "prod"

# Configuración del servicio de procesamiento (Audio-Processing-Service)
gin_mode = "release"
service_max_attempts = 5
service_timeout = 2
youtube_api_key = "TU_API_KEY_AQUI"  # ¡Acá va tu API key de YouTube!
oauth2_enabled = "true"
container_port = 8080 # Puerto del Audio-Processing-Service
secret_name = "butakero-audio-processor-secrets" # Nombre del secreto ajustado

# Tags para recursos AWS (para que todo quede ordenadito)
alb_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
ecs_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
storage_s3_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
dynamodb_table_operations_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
dynamodb_table_songs_tag = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
sqs_queue_tag = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
ecr_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
networking_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
cloudwatch_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
iam_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
security_group_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
sm_tags = {
  Project     = "music-downloader"
  Environment = "production"
  Service     = "audio-processor"
}
```


#### Configuración
1. Movete al directorio donde se encuentra los archivos .tfs:
```bash
microservices/audio_processor/song-downloader-infra
```

2. Edita el archivo [terraform.example.tfvars](../audio_processor/song-downloader-infra/terraform.example.tfvars) con tus configuraciones:
```hcl
# Configuración básica
aws_region = "us-east-1"
project_name = "butakero-music-download"
environment = "prod"

# Configuración del servicio
gin_mode = "release"
service_max_attempts = 5
service_timeout = 2
youtube_api_key = "TU_API_KEY_AQUI"  # Modifica esto con tu API key de YouTube
oauth2_enabled = "true"
container_port = 8080
secret_name = "butakero-audio-service-secrets"

# Tags para recursos AWS
alb_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
ecs_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
storage_s3_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
dynamodb_table_operations_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
dynamodb_table_songs_tag = {
  Project     = "music-downloader"
  Environment = "production"
}
sqs_queue_tag = {
  Project     = "music-downloader"
  Environment = "production"
}
ecr_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
networking_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
cloudwatch_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
iam_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
security_group_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
sm_tags = {
  Project     = "music-downloader"
  Environment = "production"
}
```

El archivo [providers.tf](../butakero_bot/butakero_bot_infra/providers.tf) debe tener esto (Backend para butakero_bot):

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "song-download-tf-state"
    key            = "music-bot/terraform.tfstate" # Clave específica para el butakero_bot
    region         = "us-east-1"
    dynamodb_table = "terraform_lock_table_music_bot" # Tabla específica para el butakero_bot
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}
```

Y el [terraform.example.tfvars](../butakero_bot/butakero_bot_infra/terraform.example.tfvars) de butakero_bot

```hcl
aws_region = "us-east-1"
discord_token = "TU_DISCORD_TOKEN_AQUI" // aca pone tu token de discord
command_prefix = "butakero"
container_port = 8081 # Puerto para el health check del butakero_bot, si lo tiene
```

#### Pasos de Despliegue (¡Importante: en orden!)

Como son dos microservicios que se interconectan, ¡los desplegamos en orden! Primero el que procesa el audio, y después el bot que lo usa.

1. Movete al directorio del audio_processor:
```bash
cd audio_processor/song-downloader-infra
```

2. Inicializá Terraform para este servicio:
```bash
terraform init
```

3. Verificá los cambios que se realizarán para el audio_processor y butakero_bot:
```bash
terraform plan -var-file="terraform.example.tfvars"
```

4. Aplicá la infraestructura del audio_processor primero, y una vez que se cree toda la infra de audio_processor, creas la de butakero_bot:
```bash
terraform apply -var-file="terraform.example.tfvars"
```
#### Recursos Desplegados
El despliegue, en total, creará los siguientes recursos en AWS para ambos servicios (cada uno con su propia configuración de Terraform):

- **ECS Cluster** con Fargate (generalmente uno solo, pero con dos servicios y 1 una tarea cada uno).
- **Application Load Balancer** (al menos uno para el audio_processor)
- **S3 Bucket** para almacenamiento de archivos (se comparte entre los servicios).
- **DynamoDB Tables** para operaciones y canciones (se comparten).
- **SQS Queue** para mensajería entre servicios (también se comparte).
- **Dos ECR Repositories** (uno para la imagen de cada microservicio).
- **CloudWatch** para logs y monitoreo (para ambos servicios).
- **IAM Roles**  y políticas necesarias (específicas para cada servicio, con los permisos justos).
- **Security Groups** para control de acceso (específicos para cada servicio).
- **VPC** y recursos de networking (compartidos).
- **Secrets Manager** para gestión de secretos.

#### Destruir la Infraestructura

Para eliminar todos los recursos creados, ¡tenés que hacerlo en orden inverso al despliegue! Primero el bot, y después el procesador de audio, así evitás problemas de dependencias.

1. Movete al directorio del butakero_bot:
```bash
cd butakero_bot/butakero_bot_infra
```

2. Destruí la infraestructura del bot:
```bash
terraform destroy -var-file="terraform.example.tfvars"
```

3. Volvé al directorio del audio_processor:
```bash
cd ../../microservices/audio_processor/song-downloader-infra
```

4. Destruí la infraestructura del audio_processor:
```bash
terraform destroy -var-file="terraform.example.tfvars"
```

## Endpoints del API

Bueno, una vez que tenés todo levantado, ¿cómo le hablás a la API? Estos endpoints son los que expone nuestro Audio-Processing-Service, que es el que hace el laburo pesado. El Butakero-Bot-Service se encarga de hablar con Discord y, a su vez, le hace pedidos a este Audio-Processing-Service para procesar la música o consultar el estado.


### 1. Health Check Del audio-Processing-Service
- **Método**: GET
- **Endpoint**: `/api/v1/health`
- **Descripción**: Este es para ver si el Audio-Processing-Service ya levantó y está listo para recibir pedidos. ¡Es lo primero que chequeamos al desplegar!

Ejemplo de solicitud:
```bash
curl -X GET "http://localhost:8080/api/v1/health"
```

Respuesta, la respuesta varia dependiendo del ambiente en el que estés:
```json
{
    "environment": "local",
    "services": {
        "kafka": {
            "metadata": {
                "kafka": {
                    "brokers": [
                        {
                            "address": "kafka:29092",
                            "is_leader": false
                        }
                    ]
                }
            },
            "status": "saludable"
        },
        "mongo_db": {
            "metadata": {
                "mongo": {
                    "replica_set_status": {
                        "role": "PRIMARY",
                        "health": 1,
                        "members": 1,
                        "last_election": "2025-05-23T18:44:21Z",
                        "replica_set_id": "rs0",
                        "sync_status": ""
                    },
                    "version": "8.0.9",
                    "connections": {
                        "active": 3,
                        "available": 838854,
                        "current": 6,
                        "rejected": 0
                    },
                    "performance": {
                        "latency_ms": 32,
                        "ops_per_sec": 211,
                        "memory_usage_mb": 190
                    }
                }
            },
            "status": "saludable"
        }
    },
    "status": "Todos los servicios son saludables"
}
```

### 2. Consultar Canción por ID en la Base de Datos

- **Método**: GET
- **Endpoint**: `/api/v1/media/`
- **Query Params**:
  - `video_id`: El ID del video de YouTube (ej: SRXH9AbT280).
- **Descripción**:  Este endpoint te permite chequear si ya tenemos un audio descargado para un video de YouTube específico. Si lo encuentra en la base de datos, te devuelve la información del audio y sus metadatos. ¡Esto es para evitar descargar lo mismo varias veces!

Ejemplo de solicitud:

```bash
curl -X GET "http://localhost:8080/api/v1/media/?video_id=k2qgadSvNyU"
```
Respuesta (ejemplo si existe):
```json
{
    "data": {
        "video_id": "k2qgadSvNyU",
        "title_lower": "dua lipa  new rules official music video",
        "status": "success",
        "message": "Iniciando descarga de la cancion",
        "metadata": {
            "title": "Dua Lipa - New Rules (Official Music Video)",
            "duration_ms": 225000,
            "url": "https://youtube.com/watch?v=k2qgadSvNyU",
            "thumbnail_url": "https://i.ytimg.com/vi/k2qgadSvNyU/default.jpg",
            "platform": "YouTube"
        },
        "file_data": {
            "file_path": "/home/enano/ButakeroMusicBotGo/microservices/audio_processor/audio-files/audio/dua lipa  new rules official music video.dca",
            "file_size": "2.70MB",
            "file_type": "audio/dca"
        },
        "processing_date": "2025-05-16T23:06:17.533Z",
        "success": true,
        "attempts": 1,
        "failures": 0,
        "created_at": "2025-05-16T23:06:17.533Z",
        "updated_at": "2025-05-16T23:06:27.597Z",
        "play_count": 0
    },
    "success": true
}
```

Respuesta (ejemplo si no existe):
```json
{
    "error": {
        "code": "media_not_found",
        "message": "Media no encontrado",
        "video_id": "k2qgadSvNyU1"
    },
    "success": false
}
```

### 3. Buscar Canciones por Titulo en la Base de Datos
- **Método**: GET
- **Endpoint**: `/api/v1/media/search`
- **Query Params**:
  - `title`: El título (o parte del título) de la canción que buscás.
- **Descripción**: Con este endpoint, podés buscar audios que ya hayamos descargado y que coincidan con un título. Te devuelve una lista de canciones con esa coincidencia, ideal para cuando el bot sugiere canciones.

Ejemplo de solicitud:
```bash
curl -X GET "http://localhost:8080/api/v1/media/search?title=dua lipa"
```

Respuesta (ejemplo):
```json
{
    "data": [
        {
            "video_id": "k2qgadSvNyU",
            "title_lower": "dua lipa  new rules official music video",
            "status": "success",
            "message": "Iniciando descarga de la cancion",
            "metadata": {
                "title": "Dua Lipa - New Rules (Official Music Video)",
                "duration_ms": 225000,
                "url": "https://youtube.com/watch?v=k2qgadSvNyU",
                "thumbnail_url": "https://i.ytimg.com/vi/k2qgadSvNyU/default.jpg",
                "platform": "YouTube"
            },
            "file_data": {
                "file_path": "/home/enano/ButakeroMusicBotGo/microservices/audio_processor/audio-files/audio/dua lipa  new rules official music video.dca",
                "file_size": "2.70MB",
                "file_type": "audio/dca"
            },
            "processing_date": "2025-05-16T23:06:17.533Z",
            "success": true,
            "attempts": 1,
            "failures": 0,
            "created_at": "2025-05-16T23:06:17.533Z",
            "updated_at": "2025-05-16T23:06:27.597Z",
            "play_count": 0
        }
    ],
    "success": true
}
```

## Notas sobre la descarga y monitoreo de canciones (¡Asíncrono y con colas!):
La descarga de canciones NO se inicia directamente a través de un endpoint HTTP. En nuestro diseño, la solicitud para descargar una canción se maneja de forma asíncrona:

## Explicación de diagramas de Arquitectura

¡Ahora sí, lo más lindo: los diagramas de arquitectura! Te cuento cómo lo pense y cómo se relacionan con nuestros dos microservicios.

## Diagrama de Secuencia

![image](/images/diagrama-de-secuencia-bot.png)

1. **Cliente (Usuario)**: Inicia la solicitud de descarga de audio a través de Discord, mandando el nombre de la canción o la URL al **Butakero-Bot-Service**.
2. **El Butakero-Bot-Service** (nuestro bot de Discord) recibe el comando de descarga.
    - Primero, puede hacer una **consulta HTTP al Audio-Processing-Service** (a /api/v1/media/?video_id= o /api/v1/media/search?title=) para ver si la canción ya está en la base de datos.
    - Si no la encuentra, envía un mensaje con la solicitud de descarga a la **Cola de Pedidos** (ej: bot.download.requests en Kafka/SQS).
3. **Cola de Pedidos (Kafka/SQS)**: El mensaje con la solicitud de descarga se encola acá.
4. **Audio-Processing-Service**: Está escuchando la Cola de Pedidos, consume el mensaje, busca el ID del video en YouTube (vía **YouTube API**), y luego descarga y procesa el audio.
5. **YouTube API**: Proporciona el ID del video y sus detalles (metadata) al Audio-Processing-Service.
6. **Audio-Processing-Service**: Guarda el audio en **Amazon S3** o **local** y actualiza el estado de la operación y metadatos en **DynamoDB** o **MongoDB**.
7. **Audio-Processing-Service**: Una vez que termina de procesar, publica un mensaje con el resultado (éxito/error, metadatos) en la **Cola de Estado/Respuestas** (ej: bot.download.status).
8. **Cola de Estado/Respuestas (Kafka/SQS)**: El mensaje con el resultado de la operación se encola acá.
9. **Butakero-Bot-Service**: Está escuchando la Cola de Estado/Respuestas, consume el mensaje, y notifica al **Cliente (Usuario)** a través de Discord el resultado de la descarga.

Este enfoque asíncrono con colas esta bueno porque vos recibís una respuesta inmediata del bot, y nuestro sistema se encarga de todo por detrás, sin bloquear la interacción del usuario.



## Pruebas

Para ejecutar las pruebas unitarias y de integracion del proyecto, podes correr el [script de pruebas que se encuentra en la raiz del proyecto](../Makefile). Este script ejecuta las pruebas de ambos microservicios, el bot y el audio-processor.:

```bash
make test
```


## Explicación de diagramas de Arquitectura

## Diagrama de Secuencia

![image](/images/diagrama-de-secuencia-bot.png))

El diagrama de secuencia ilustra el flujo de interacción entre los diferentes componentes del microservicio durante el proceso de descarga y procesamiento de audio. A continuación, se describen los pasos clave:

1. **Cliente (Usuario)**: Inicia la solicitud de descarga de audio a través de Discord, mandando el nombre de la canción o la URL al Butakero-Bot-Service.
2. **Butakero-Bot-Service**: Recibe el comando.
    - Primero, puede hacer una consulta HTTP al Audio-Processing-Service (a /api/v1/media/?video_id= o /api/v1/media/search?title=) para ver si la canción ya está en la base de datos.
    - Si no la encuentra, envía un mensaje con la solicitud de descarga a la Cola de Pedidos (ej: bot.download.requests en Kafka/SQS).
3. **Cola de Pedidos (Kafka/SQS)**: El mensaje con la solicitud de descarga se encola acá.
4. **Audio-Processing-Service**: Está escuchando la Cola de Pedidos, consume el mensaje, busca el ID del video en YouTube (vía YouTube API), y luego descarga y procesa el audio.

Este enfoque asíncrono asegura que el usuario reciba una respuesta inmediata, mejorando la experiencia del usuario.

# Arquitectura de la Aplicación en AWS ECS

![image](/images/arquitectura-microservices-aws.png)

Acá te muestro cómo lo armamos en AWS ECS. La idea es que este pensado para bancarse la carga y ser robusto.

## Componentes de la Arquitectura

### 1. VPC (Virtual Private Cloud)
Todo corre dentro de una VPC para asegurar que los recursos estén aislados y podamos aplicar reglas de seguridad específicas. Esto nos permite controlar el tráfico y proteger los servicios.

### 2. Public Subnet 
Ambos microservicios (Audio-Processing-Service y Butakero-Bot-Service) se despliegan en una subred pública. Esta ubicación es estratégica para permitirles:
   - Conectarse a servicios externos como las APIs de YouTube.
   - Para el Butakero-Bot-Service, recibir los eventos de Discord (ya que Discord se comunica con la IP pública del bot).

### 3. Application Load Balancer (ALB)
El ALB es clave acá. Recibe el tráfico HTTP destinado al Audio-Processing-Service (por ejemplo, para el health check /api/v1/health, o las consultas de canciones como /api/v1/media/*). **El Butakero-Bot-Service también se comunica con el Audio-Processing-Service a través de este mismo ALB**, haciendo que sea el punto central para las interacciones HTTP entre ellos.

- **Reglas de Ingress (Entrada)**: El tráfico que llega al ALB está restringido por Security Groups. Solo se permite tráfico entrante desde:
  - Los **Security Groups de las tareas de ECS del Butakero-Bot-Service**: Esto asegura que solo nuestro bot pueda realizar consultas HTTP al Audio-Processing-Service a través del ALB.
  - Tu **IP pública específica**: Para que vos, desde tu máquina, puedas probar los endpoints directamente.
- **Reglas de Egress (Salida)**: El tráfico saliente desde el ALB se dirige al **Security Group de las tareas del Audio-Processing-Service**.

### 4. ECS Cluster y Fargate
Acá es donde viven nuestros dos microservicios: el **Butakero-Bot-Service** y el **Audio-Processing-Service**. Usamos Fargate, ¡y eso es lo más! No tenemos que gestionar la infraestructura de los contenedores, Fargate se encarga de las máquinas. Nosotros nos concentramos en que nuestro código ande.

### 5. ECS Tasks
Estas tareas son donde realmente corre nuestro código.
  - Las tareas del Butakero-Bot-Service se conectan a Discord, y para las solicitudes de descarga, interactúan con las colas **SQS**. Para consultas de canciones ya existentes, ¡hacen llamadas HTTP al ALB del Audio-Processing-Service!
  - Las tareas del Audio-Processing-Service escuchan la **Cola SQS de Pedidos** y publican en la **Cola SQS de Estado**. También, exponen los endpoints HTTP para consultas de media que son accedidos por el ALB.
  - **Container**: Cada Task corre su propio Container con la imagen de Docker de cada servicio.
  - **Interacción con S3**: Nuestras tareas ECS pueden acceder a S3 para subir o descargar objetos. Por ejemplo, ahí guardamos los audios procesados (.dca, etc.). Hablan con la API de S3 para manejar esos archivos.
  - **Interacción con DynamoDB**: También, nuestras tareas se conectan a DynamoDB para registrar el estado de la aplicación, como el seguimiento de operaciones o los metadatos de las canciones. DynamoDB es re rápido, ideal para lo que necesitamos.

### 6. SQS Queues
Acá tenemos dos colas principales: una para las **solicitudes de descarga** (donde el bot publica) y otra para los **estados de descarga** (donde el procesador publica y el bot consume). Esto asegura la comunicación asíncrona y robusta.

### 7. Secrets Manager, IAM, ECR
Estos son servicios de soporte cruciales.
  - **Secrets Manager**: Para guardar de forma segura tokens de Discord, API Keys de YouTube y otras credenciales que nuestros servicios necesitan.
  - **IAM (Identity and Access Management)**: Proporciona los roles y permisos necesarios para que las tareas de ECS interactúen con otros servicios de AWS (SQS, S3, DynamoDB, Secrets Manager). Cada tarea de ECS tiene asociado un **IAM Role** específico con los permisos mínimos necesarios.
  - **ECR (Elastic Container Registry)**: Nuestro repositorio privado de Docker donde guardamos las imágenes de nuestros microservicios antes de desplegarlas en ECS.

### 8. CloudWatch y Auto Scaling Group (en el contexto de Fargate):
  - El monitoreo está a cargo de CloudWatch. Configuramos métricas clave, como el uso de CPU y memoria en las tareas ECS (Fargate).
  - Las **políticas de Auto Scaling** (que CloudWatch activa) permiten escalar horizontalmente las tareas de ECS de cada servicio de forma independiente, lanzando más tareas si la demanda aumenta y reduciéndolas cuando no son necesarias. Aunque el diagrama muestra "Auto Scaling Group" e "Instances", en Fargate esto se refiere a la escalabilidad de las tareas, no a la gestión de instancias EC2 subyacentes.

## Flujo de Tráfico

1. **Clientes (Discord)**: Los usuarios interactúan con el **Butakero-Bot-Service** a través de Discord.
2. **Butakero-Bot-Service (ECS Task)**: Este microservicio procesa los comandos de Discord.
   - Para consultas de canciones (si ya existen en la DB), hace llamadas HTTP **al ALB que apunta al Audio-Processing-Service**.
   - Para solicitudes de descarga, envía un mensaje a la **Cola SQS de Pedidos**.
   - Escucha la **Cola SQS de Estado** para recibir las respuestas del Audio-Processing-Service y notificar al usuario.
3. **Audio-Processing-Service (ECS Task)**: Al consumir un mensaje de la cola de pedidos, busca el video en YouTube (usando la YouTube API), descarga, procesa el audio y lo sube a S3. Luego, actualiza el estado de la operación en DynamoDB y publica el resultado en la **Cola SQS de Estado**.
4. **ALB (para Audio-Processing-Service)**: Si una aplicación externa (o el mismo bot para consultas específicas) necesita interactuar directamente con la API HTTP del Audio-Processing-Service (ej. /api/v1/health), el ALB dirige el tráfico a las tareas del Audio-Processing-Service.
5. **S3 y DynamoDB**: Ambos servicios interactúan con estas bases de datos para guardar archivos y registrar estados.
6. **CloudWatch**: Monitorea el rendimiento de todas las tareas y activa las políticas de escalado si hace falta.
---

## Arquitectura en Kubernetes

![image](/images/arquitectura-en-k8s.png)

Nuestra aplicación está desplegada en Kubernetes utilizando una arquitectura de microservicios que consta de varios componentes clave:

### Componentes Core

1. **Butakero-Bot-Service**: Este es nuestro primer microservicio. Es el punto de entrada para los comandos de Discord. Está desplegado como un servicio de Kubernetes (Tipo ClusterIP si lo exponés con un Ingress/LoadBalancer de K8s, pero no un ALB externo). Se comunica con el Audio-Processing-Service para consultas HTTP y con Kafka para iniciar/monitorear descargas.
2. **Audio-Processing-Service**: Nuestro segundo microservicio, el que hace el laburo pesado de procesamiento. Está desplegado como otro servicio de Kubernetes (también ClusterIP). Recibe pedidos del Butakero-Bot-Service (vía HTTP para consultas y vía Kafka para las descargas).
3. **Capa de Aplicación**: En el diagrama, estas son las instancias (pods) de nuestros dos backends en tiempo real (Butakero-Bot-Service y Audio-Processing-Service). Están como Deployments para asegurar que todo funcione como debe y puedan escalar horizontalmente según la demanda.
    - **Comunicación interna**: La comunicación entre Butakero-Bot-Service y Audio-Processing-Service dentro del clúster se hace usando los nombres de servicio de Kubernetes. Por ejemplo, el bot se conecta al procesador con una URL como http://audio-processing-service.backend.svc.cluster.local:8080 (donde audio-processing-service es el nombre del servicio, backend es el namespace, svc es el tipo de recurso y cluster.local es el dominio del clúster). Esto evita la necesidad de Load Balancers externos entre ellos.
4. **Procesamiento de Mensajes (Kafka y MongoDB)**: Son el corazón de la comunicación asíncrona entre nuestros servicios y donde guardamos la data. Usamos PVC (Persistent Volume Claims) para que los datos no se pierdan si un pod se reinicia. Los gestionamos usando operadores de Kubernetes, así no nos hacemos un quilombo con la configuración manual.

### Operadores de Kubernetes

El sistema usa operadores especializados para gestionar las aplicaciones que necesitan mantener estado.

- **Operador Strimzi**: Es el que gestiona nuestros clusters de Kafka.
- **Operador MongoDB**: Maneja los despliegues y escalado de MongoDB

### Stack de Monitoreo

- **Prometheus**: Recolecta métricas de todos nuestros servicios y del clúster.
- **Grafana**: Visualiza todos esos datos de monitoreo en dashboards re lindos y fáciles de entender.
- **Colectores de Métricas**:
  - Node exporter para las métricas del sistema operativo de los nodos.
  - Service monitor para recolectar métricas de nuestros microservicios directamente.
  - Cadvisor para recolectar métricas de los contenedores.

### Almacenamiento y Gestión de Estado

- Aseguramos el almacenamiento persistente usando PVC para:
  - Los datos de MongoDB.
  - Logs de Kafka
  - Los metadatos de operaciones
  - El almacenamiento de las canciones procesadas.

### Seguridad

- **Cert-manager**: Nos ayuda a gestionar los certificados SSL/TLS automáticamente para la comunicación segura.
- Tenemos comunicación segura entre servicios usando TLS.
- Configuramos políticas de red para controlar la comunicación interna entre los pods.


### Escalabilidad y Confiabilidad

- Todos nuestros componentes importantes están replicados para asegurar alta disponibilidad:
  - Clusters de Kafka con múltiples réplicas.
  - ReplicaSets de MongoDB
  - Múltiples instancias de nuestros microservicios (Butakero-Bot-Service y Audio-Processing-Service).