# ButakeroMusicBot

**ButakeroMusicBot** es un microservicio que permite la descarga, procesamiento y subida de audios desde videos de YouTube a un bucket de Amazon S3. Está diseñado para funcionar en la nube, utilizando DynamoDB para guardar el estado de las operaciones y los metadatos de las canciones procesadas.

## Funcionalidades

- Búsqueda de videos en YouTube por nombre o URL.
- Descarga y procesamiento de audio.
- Subida de archivos de audio a Amazon S3.
- Registro de operaciones y metadatos en DynamoDB.
- Sistema de reintentos en caso de fallos en el procesamiento.

## Requisitos

Este microservicio depende de los siguientes servicios externos:

- **Amazon S3**: Para almacenar los archivos de audio procesados.
- **Amazon DynamoDB**: Para registrar los metadatos de las canciones y los resultados de las operaciones.
- **YouTube API**: Para obtener los detalles de los videos de YouTube.

En el futuro, se implementará una versión que funcionará en memoria, eliminando la dependencia de servicios externos, lo que permitirá correr el microservicio sin conexión a AWS.

## Futuras Implementaciones

Próximamente se agregarán las siguientes funcionalidades:

- **WebSockets**: Para notificar al usuario el estado de sus operaciones en tiempo real.
- **Implementación en memoria**: Una versión que no dependa de servicios externos como AWS, permitiendo la ejecución local.

## Ejecución

Para ejecutar el microservicio:

1. Clonar el repositorio.
2. Configurar las variables de entorno.
3. Ejecutar el servicio utilizando `go run`:

```bash
go run cmd/main.go
```

El servicio correrá en el puerto 8080 y estará listo para recibir solicitudes de procesamiento de canciones.

## Endpoints del API

### 1. Iniciar el procesamiento de una canción

- **Método**: POST
- **Endpoint**: `/api/audio/start`
- **Query Params**:
  - `song`: El título de la canción o la URL del video de YouTube.
- **Descripción**: Este endpoint inicia el procesamiento de la canción. Se puede enviar el nombre o la URL de la canción en el parámetro song. La API buscará el video en YouTube, descargará el audio, lo procesará y lo subirá a S3.

Ejemplo de solicitud:

```bash
curl -X POST "http://localhost:8080/api/audio/start?song=Never+Gonna+Give+You+Up"
```

Respuesta:

```json
{
  "operation_id": "unique-operation-id",
  "song_id": "unique-song-id"
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
curl -X GET "http://localhost:8080/api/audio/status?operation_id=unique-operation-id&song_id=unique-song-id"
```

Respuesta (ejemplo de operación en curso):

```json
{
  "status": {
    "id": "unique-operation-id",
    "song_id": "unique-song-id",
    "status": "iniciando",
    "message": "",
    "data": "",
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
    "id": "unique-operation-id",
    "song_id": "unique-song-id",
    "status": "success",
    "message": "Procesamiento exitoso",
    "data": "Archivo guardado en S3: s3://bucket/Rick Astley - Never Gonna Give You Up (Official Music Video).dca",
    "processing_date": "2024-10-02T14:21:41-03:00",
    "success": true,
    "attempts": 1,
    "failures": 0
  }
}
```

Respuesta (ejemplo de operación fallida):

```json
{
  "operation_id": "unique-operation-id",
  "status": "failed",
  "error": "Descripción del error ocurrido"
}
```

## Detalles Técnicos

### Configuración del microservicio

El microservicio se configura utilizando un archivo de configuración (cfg) que se construye con los siguientes parámetros:

```go
cfg := config.Config{
    MaxAttempts:           3,
    Timeout:               4 * time.Minute,
    BucketName:            os.Getenv("BUCKET_NAME"),
    Region:                os.Getenv("REGION"),
    YouTubeApiKey:         os.Getenv("YOUTUBE_API_KEY"),
    SongsTable:            os.Getenv("DYNAMODB_TABLE_NAME_SONGS"),
    OperationResultsTable: os.Getenv("DYNAMODB_TABLE_NAME_OPERATION"),
    AccessKey:             os.Getenv("ACCESS_KEY"),
    SecretKey:             os.Getenv("SECRET_KEY"),
    Environment:           os.Getenv("GIN_MODE"), 
}
```

Asegurate de tener configuradas las siguientes variables de entorno:

- `BUCKET_NAME`: El nombre del bucket en S3 donde se almacenarán los audios procesados.
- `REGION`: La región de AWS.
- `YOUTUBE_API_KEY`: La clave de API de YouTube.
- `DYNAMODB_TABLE_NAME_SONGS`: El nombre de la tabla de DynamoDB para los metadatos de canciones.
- `DYNAMODB_TABLE_NAME_OPERATION`: El nombre de la tabla de DynamoDB para los resultados de operaciones.
- `ACCESS_KEY` y `SECRET_KEY` : Credenciales de AWS necesarias para acceder a los servicios.
- `GIN_MODE`: Modo de ejecuccion de la applicacion de Gin. Puede ser `release` o `debug`, por default esta en `default`.

## Pruebas

Para ejecutar las pruebas unitarias y de integracion del proyecto, podes correr:

```bash
go test ./...
```


## Explicación de los Diagramas de Secuencia y Arquitectura

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