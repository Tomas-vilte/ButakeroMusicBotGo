# ButakeroMusicBot

**ButakeroMusicBot** es un microservicio diseñado para la descarga, procesamiento y subida de audios desde videos de YouTube a Amazon S3. Utiliza tecnologías modernas y está preparado para ser desplegado en la nube, Kubernetes o localmente con Docker Compose.


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

El microservicio puede ser desplegado de distintas maneras, dependiendo del entorno en el que se desee correr. A continuación, se detallan las tres formas principales de despliegue:

### 1. Docker Compose

El despliegue con Docker Compose levanta los siguientes servicios:

- Zookeeper
- Kafka
- MongoDB
- Backend (usando el Dockerfile del proyecto)

### Configuración Inicial

1. Docker y Docker Compose instalados en tu máquina.

2. Crea un archivo llamado `test_local.env`, en el directorio donde se encuentra el [docker compose](./docker-compose.yml). Este archivo tiene que tener estas variables necesarias para ejecutar el servicio localmente. Lo principal que necesitas modificar es la variable `YOUTUBE_API_KEY`:

```env
# Environment
ENVIRONMENT=local
LOCAL_STORAGE_PATH=data/audio-files
SERVICE_MAX_ATTEMPTS=7
SERVICE_TIMEOUT=2
YOUTUBE_API_KEY=TU_API_KEY_AQUI  # Modificar esto con tu API key
GIN_MODE=release

# Kafka Configuration
KAFKA_TOPIC=notification
KAFKA_BROKERS=kafka:29092
KAFKA_ENABLE_TLS=false

# Authentication
OAUTH2=false

# MongoDB Configuration
MONGO_USER=root
MONGO_PASSWORD=root
MONGO_PORT=27017
MONGO_HOST=mongodb
MONGO_DATABASE=audio_service_db
MONGO_COLLECTION_SONGS=Songs
MONGO_COLLECTION_OPERATIONS=Operations
MONGO_ENABLE_TLS=false
MONGO_REPLICA_SET_NAME=rs0

# YouTube Cookies
COOKIES_YOUTUBE=/root/yt-cookies.txt
```

#### Obtener YouTube API Key

Para obtener tu YouTube API Key, sigue estos pasos:

1. Ve a la [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto o selecciona uno existente
3. Habilita la YouTube Data API v3 para tu proyecto
4. Ve a "Credenciales"
5. Crea una nueva credencial de tipo "API Key"
6. Copia la API Key generada y colócala en la variable `YOUTUBE_API_KEY` del archivo `test_local.env`

3. Script de cookies de YouTube (opcional): Si encontras problemas con YouTube, como el error "Sign in to confirm your age", podes generar un archivo de cookies para evitar el bloqueo. Para ello, tenes que ejecutar previamente el script [filter_youtube_cookies.sh](filter_youtube_cookies.sh) que generará el archivo `yt-cookies.txt`.

```bash
bash filter_youtube_cookies.sh
```

#### Ejecutar con Docker Compose
El proyecto incluye un [docker-compose.yml](./docker-compose.yml) que levanta todos los servicios necesarios:
- Zookeeper
- Kafka
- MongoDB
- Backend (usando el Dockerfile del proyecto)

Para iniciar todos los servicios:
```bash
docker compose up -d
```

Para ver los logs de los servicios:
```bash
# Ver todos los logs
docker compose logs -f

# Ver logs de un servicio específico
docker compose logs -f backend-application
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
Antes de ejecutar el despliegue, debes configurar las credenciales en el archivo [secret.yaml](/k8s/bases/backend/download-service/secret.yaml):
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
```

#### Obtener Cookies de YouTube (Opcional)
Para evitar errores de autorización en YouTube ("Sign in to confirm your age"), puedes configurar cookies:
1. Ejecuta el script:
```bash
bash filter_youtube_cookies.sh
```

2. Copia el contenido generado al campo YT_COOKIES del secret.

#### Despliegue Automatizado

El proyecto incluye un script [deploy.sh](/deploy.sh) que automatiza todo el proceso de despliegue. Este script:

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
kubectl port-forward svc/backend-service 8080 -n backend
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
kubectl logs -f backend-application-hash -n backend
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
> Antes de seguir, es **obligatorio** generar las cookies con el script [filter_youtube_cookies.sh](filter_youtube_cookies.sh). Muchas IPs de las instancias de AWS están **baneadas o bloqueadas** por YouTube, así que vas a tener que usar cookies válidas para evitar problemas de conexión. Si vas a desplegarlo en ECS, asegurate de hacer este paso antes.

1. Crear el bucket S3:
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

3. Crear la tabla DynamoDB para el bloqueo del estado:
```bash
aws dynamodb create-table \
    --table-name terraform-lock-table \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region us-east-1
```

#### Configuración del Provider y Backend

El archivo [provider.tf](song-downloader-infra/providers.tf) debe contener:

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
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-lock-table"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}
```

#### Configuración
1. Movete al directorio donde se encuentra los archivos .tfs:
```bash
microservices/audio_processor/song-downloader-infra
```

2. Edita el archivo [terraform.example.tfvars](song-downloader-infra/terraform.example.tfvars) con tus configuraciones:
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

#### Pasos de Despliegue

1. Inicializa Terraform:
```bash
terraform init
```

2. Verifica los cambios que se realizarán:
```bash
terraform plan -var-file="terraform.example.tfvars"
```

3. Aplica la infraestructura:
```bash
terraform apply
```
#### Recursos Desplegados
El despliegue creará los siguientes recursos en AWS:

- **ECS Cluster** con Fargate
- **Application Load Balancer**
- **S3 Bucket** para almacenamiento de archivos
- **DynamoDB Tables** para operaciones y canciones
- **SQS Queue** para mensajería
- **ECR Repository** para imágenes de contenedor
- **CloudWatch** para logs y monitoreo
- **IAM Roles** y políticas necesarias
- **Security Groups** para control de acceso
- **VPC** y recursos de networking
- **Secrets Manager** para gestión de secretos

#### Destruir la Infraestructura

Para eliminar todos los recursos creados:
```bash
terraform destroy
```


## Endpoints del API

### 1. Iniciar el procesamiento de una canción

- **Método**: POST
- **Endpoint**: `/api/audio/start`
- **Query Params**:
  - `song`: El título de la canción o la URL del video de YouTube.
- **Descripción**: Este endpoint inicia el procesamiento de la canción. Se puede enviar el nombre o la URL de la canción en el parámetro song. La API buscará el video en YouTube, descargará el audio, lo procesará y lo subirá a S3.

Ejemplo de solicitud:

```bash
curl -X POST "http://localhost:8080/api/v1/audio/start?song=linkin+park+the+emptiness+machine"
```

Respuesta:

```json
{
  "operation_id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
  "song_id": "SRXH9AbT280"
}
```

### 2. Consultar el estado de una operación

- **Método**: GET
- **Endpoint**: `/api/audio/status`
- **Query Params**:
  - `operation_id`: El ID único de la operación iniciada.
  - `song_id`: El ID de la canción procesada.
- **Descripción**: Este endpoint devuelve el estado actual del procesamiento de audio utilizando el `operation_id` y el `song_id`. El estado incluye información detallada sobre la operación.

Ejemplo de solicitud:

```bash
curl -X GET "http://localhost:8080/api/v1/audio/status?operation_id=19f6c66f-26f3-4ccf-bfc7-967449a95ad4&song_id=SRXH9AbT280"
```

Respuesta (ejemplo de operación en curso):

```json
{
  "status": {
      "id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
      "sk": "SRXH9AbT280",
      "status": "iniciando",
      "message": "",
      "metadata": null,
      "file_data": null,
      "processing_date": "",
      "success": false,
      "attempts": 0,
      "failures": 0
  }
}
```

Respuesta (ejemplo de operación finalizada):

```json
{
    "status": {
      "id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
      "sk": "SRXH9AbT280",
      "status": "success",
      "message": "Procesamiento exitoso",
      "metadata": {
          "id": "63f48016-78cd-4387-99b9-c38af46e8e90",
          "video_id": "SRXH9AbT280",
          "title": "The Emptiness Machine (Official Music Video) - Linkin Park",
          "duration": "PT3M21S",
          "url_youtube": "https://youtube.com/watch?v=SRXH9AbT280",
          "thumbnail": "https://i.ytimg.com/vi/SRXH9AbT280/default.jpg",
          "platform": "Youtube"
      },
      "file_data": {
          "file_path": "audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca",
          "file_size": "1.44MB",
          "file_type": "audio/dca",
          "public_url": "file://data/audio-files/audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca"
      },
      "processing_date": "2024-12-24T05:39:58Z",
      "success": true,
      "attempts": 1,
      "failures": 0
  }
}
```

Respuesta (ejemplo de operación fallida):

```json
{
  "operation_id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
  "status": "failed",
  "error": "Descripción del error ocurrido"
}
```

## Pruebas

Para ejecutar las pruebas unitarias y de integracion del proyecto, podes correr:

```bash
go test ./...
```


## Explicación de diagramas de Arquitectura

## Diagrama de Secuencia

![image](/images/diagrama-de-secuencia-microservicio.png)

El diagrama de secuencia ilustra el flujo de interacción entre los diferentes componentes del microservicio durante el proceso de descarga y procesamiento de audio. A continuación, se describen los pasos clave:

1. **Cliente**: Inicia la solicitud de descarga de audio enviando la canción deseada al microservicio.
2. **Microservicio**: Recibe la solicitud y utiliza el servicio de YouTube para buscar el ID del video correspondiente a la canción.
3. **YouTube API**: Proporciona el ID del video y sus detalles (metadata) al microservicio.
4. **Microservicio**: Inicia una operación para el procesamiento del audio y devuelve el `operation_id` y `song_id` al cliente.
5. **Proceso Asíncrono**: En paralelo, el microservicio procesa el audio utilizando el ID de operación y los detalles obtenidos, permitiendo al cliente continuar con otras tareas sin esperar la finalización.

Este enfoque asíncrono asegura que el usuario reciba una respuesta inmediata, mejorando la experiencia del usuario.

# Arquitectura de la Aplicación en AWS ECS

![image](/images/arquitectura-microservice-aws.png)

## Componentes de la Arquitectura

### 1. VPC (Virtual Private Cloud)
Todo corre dentro de una VPC para asegurar que los recursos estén aislados y podamos aplicar reglas de seguridad específicas. Esto nos permite controlar el tráfico y proteger los servicios.

### 2. EC2 Instance
El tráfico llega primero a una instancia de EC2 que actúa como puerta de entrada. Desde aca, la aplicacion **Nuestro Bot de musica** envía requests a nuestra aplicación **Donde se encuentra la logica de procesamiento de audio**, que son redirigidas a través de un Application Load Balancer.

### 3. Application Load Balancer (ALB)
El ALB es clave en esta arquitectura. Recibe tráfico HTTP en el puerto 80 y lo distribuye a un Target Group configurado para enrutar las solicitudes a las tareas de ECS. Además, tiene configurado un **health check** que verifica cada 30 segundos el estado de las tareas para garantizar que solo las instancias saludables reciban tráfico.

### 4. ECS Cluster y Fargate
Estamos usando Fargate, lo que significa que no tenemos que gestionar la infraestructura de los contenedores. Las tareas de ECS se ejecutan dentro del clúster, sin la necesidad de manejar instancias EC2. Esto nos permite concentrarnos en el desarrollo, y Fargate se encarga del resto.

### 5. ECS Tasks
Las tareas de ECS son donde realmente se ejecuta nuestro código. Los contenedores están corriendo en el puerto 8080, y el ALB enruta el tráfico hacia este puerto desde el Target Group. Cada tarea tiene permisos para interactuar con servicios como S3 y DynamoDB a través de roles IAM configurados específicamente para esto.

#### Interacción con S3
Nuestras tareas ECS pueden acceder a S3 para subir o descargar objetos. Por ejemplo, usamos S3 para almacenar archivos multimedia por ej **.dca**. Las tareas hacen uso de la API de S3 para gestionar estos archivos.

#### Interacción con DynamoDB
Además, las tareas se conectan a DynamoDB para gestionar el estado de la aplicación, como el seguimiento de operaciones o el almacenamiento de metadatos. DynamoDB es rápido y se adapta bien a las necesidades de baja latencia de nuestra aplicación.

### 6. CloudWatch y Auto Scaling
El monitoreo está a cargo de CloudWatch. Configuramos métricas clave, como el uso de CPU y memoria en las tareas ECS. En base a estas métricas, tenemos configuradas **políticas de Auto Scaling**, que permiten escalar horizontalmente las tareas de ECS. Esto significa que si el uso de CPU o memoria supera ciertos umbrales, automáticamente se lanzan más tareas para manejar el tráfico adicional, y se reducen cuando ya no son necesarias.

#### Configuración del Auto Scaling
Configuramos el Auto Scaling usando CloudWatch como desencadenante. Cuando se alcanza un cierto umbral de CPU o memoria (por ejemplo, 75%), se activa la política que lanza nuevas tareas ECS hasta que los recursos vuelvan a estar en niveles normales. Esto asegura que nuestra aplicación se mantenga eficiente sin desperdiciar recursos.

### 7. IAM Roles y Seguridad
Cada tarea de ECS tiene asociado un **IAM Role** que le permite acceder a servicios específicos como S3 y DynamoDB, pero sin dar permisos innecesarios. Estos roles están configurados con permisos mínimos para garantizar la seguridad. Por otro lado, usamos **Security Groups** para controlar el tráfico entrante y saliente en la VPC, asegurando que solo el tráfico autorizado llegue a los contenedores.

## Flujo de Tráfico

1. **Applicacion bot de musica**: Envía una requests desde un cliente EC2.
2. **ALB**: La solicitud llega al ALB, que se encarga de dirigir el tráfico al Target Group.
3. **Target Group**: Este grupo enruta el tráfico a las tareas de ECS corriendo en Fargate.
4. **ECS Tasks**: Las tareas procesan la solicitud y, si necesitan, acceden a S3 y DynamoDB para manejar los datos.
5. **CloudWatch**: Monitorea el rendimiento y activa políticas de escalado si se detectan problemas de capacidad.
---

## Arquitectura en Kubernetes

![image](/images/arquitectura-k8s.png)

Nuestra aplicación está desplegada en Kubernetes utilizando una arquitectura de microservicios que consta de varios componentes clave:

### Componentes Core

1. **Servicio Backend**
   - Punto de entrada para todas las solicitudes de clientes
   - Desplegado como un servicio de Kubernetes con capacidades de balanceo de carga

2. **Capa de Aplicación**
   - Contiene los nodos de Backend en tiempo real
   - Desplegado como StatefulSet para mantener el orden de procesamiento
   - Escala horizontalmente según la demanda

3. **Procesamiento de Mensajes**
   - Servicios de Kafka y MongoDB para el manejo y almacenamiento de mensajes
   - Utiliza PVC (Persistent Volume Claims) para persistencia de datos
   - Desplegado usando operadores de Kubernetes para gestión automatizada

### Operadores de Kubernetes

El sistema utiliza operadores especializados para gestionar aplicaciones con estado:

- **Operador Strimzi**: Gestiona los clusters de Kafka
- **Operador MongoDB**: Maneja los despliegues y escalado de MongoDB

### Stack de Monitoreo

- **Prometheus**: Recolecta métricas de los servicios
- **Grafana**: Visualiza datos de monitoreo
- **Colectores de Métricas**:
  - Node exporter para métricas del sistema
  - Service monitor para métricas de aplicación

### Almacenamiento y Gestión de Estado

- Almacenamiento persistente usando PVC para:
  - Datos de MongoDB
  - Logs de Kafka
  - Metadatos de operaciones
  - Almacenamiento de canciones

### Seguridad

- **Cert-manager**: Gestión de certificados TLS
- Comunicación segura entre servicios usando TLS
- Políticas de red para control de comunicación interna

### Escalabilidad y Confiabilidad

- Componentes replicados para alta disponibilidad:
  - Clusters de Kafka con múltiples réplicas
  - ReplicaSets de MongoDB
  - Múltiples instancias del Backend en tiempo real
