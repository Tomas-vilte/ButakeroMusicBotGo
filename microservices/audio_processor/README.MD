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
  "operationID": "unique-operation-id",
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
}
```

Asegurate de tener configuradas las siguientes variables de entorno:

- `BUCKET_NAME`: El nombre del bucket en S3 donde se almacenarán los audios procesados.
- `REGION`: La región de AWS.
- `YOUTUBE_API_KEY`: La clave de API de YouTube.
- `DYNAMODB_TABLE_NAME_SONGS`: El nombre de la tabla de DynamoDB para los metadatos de canciones.
- `DYNAMODB_TABLE_NAME_OPERATION`: El nombre de la tabla de DynamoDB para los resultados de operaciones.

## Pruebas

Para ejecutar las pruebas unitarias y de integracion del proyecto, podes correr:

```bash
go test ./...
```


## Explicación de los Diagramas de Secuencia y Arquitectura

## Diagrama de Secuencia

![image](/docs/diagrama-de-secuencia-microservicio.png)

El diagrama de secuencia ilustra el flujo de interacción entre los diferentes componentes del microservicio durante el proceso de descarga y procesamiento de audio. A continuación, se describen los pasos clave:

1. **Cliente**: Inicia la solicitud de descarga de audio enviando la canción deseada al microservicio.
2. **Microservicio**: Recibe la solicitud y utiliza el servicio de YouTube para buscar el ID del video correspondiente a la canción.
3. **YouTube API**: Proporciona el ID del video y sus detalles (metadata) al microservicio.
4. **Microservicio**: Inicia una operación para el procesamiento del audio y devuelve el `operationID` y `songID` al cliente.
5. **Proceso Asíncrono**: En paralelo, el microservicio procesa el audio utilizando el ID de operación y los detalles obtenidos, permitiendo al cliente continuar con otras tareas sin esperar la finalización.

Este enfoque asíncrono asegura que el usuario reciba una respuesta inmediata, mejorando la experiencia del usuario.

## Diagrama de Arquitectura

![image](/docs/diagram-arquitectura-microservicio.png)

El diagrama de arquitectura presenta la infraestructura general del microservicio, destacando los componentes principales y su interconexión. A continuación, se describen los elementos clave:

1. **Load Balancer**: Distribuye las solicitudes entrantes entre las instancias del microservicio, garantizando una carga equilibrada y alta disponibilidad.
2. **Microservicio**: El núcleo de la aplicación que maneja las solicitudes de descarga de audio. Se comunica con:
   - **YouTube API**: Para buscar y obtener los detalles del video.
   - **DynamoDB**: Para almacenar resultados de operaciones y metadatos relacionados con las canciones.
   - **AWS CloudWatch**: Para la monitorización y registro de logs, permitiendo un seguimiento eficaz del rendimiento y posibles errores.
    - **AWS S3**: Para almacenar las canciones para no tener que volver a descargarlas, estos archivos se guardan como `.dca`.

Este diagrama ilustra cómo se interrelacionan los distintos componentes dentro de AWS, ofreciendo un marco robusto y escalable para el microservicio.