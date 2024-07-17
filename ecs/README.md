# Music Bot Feature: Audio Processing

## Descripción

Este documento describe la arquitectura y funcionamiento de la feature que permite procesar canciones que no están disponibles en S3. Esta implementación está desacoplada del bot principal, asegurando una separación de responsabilidades.


## Funcionamiento General

Cuando un usuario solicita una canción que no existe en S3, se activa un flujo de trabajo que incluye la descarga del archivo desde YouTube, su procesamiento y la subida del archivo resultante a S3.

1. **Detección de Canción No Disponible**: Si la canción no está en S3, se inicia el flujo.
2. **Descarga desde YouTube**: Se descarga la canción en formato .m4a.
3. **Subida a S3**: El archivo .m4a se sube a un bucket en S3.
4. **Activación de Lambda**: Al subir el archivo, una función de AWS Lambda se activa.
5. **Recolección de Información**: La Lambda recolecta información del archivo recién subido y actualiza el estado en DynamoDB a `PROCESSING`.
6. **Verificación de Estado**: Antes de enviar el job a ECS, se verifica en DynamoDB si el registro de la canción ya ha sido procesado. Esto ayuda a evitar procesar la misma canción múltiples veces.
7. **Ejecución en ECS**: Si la canción no ha sido procesada, se envía la información a un servicio de AWS ECS que ejecuta el procesamiento del archivo, convirtiéndolo a formato .dca.
8. **Subida del Archivo Procesado**: El archivo .dca se sube de nuevo a S3.
9. **Notificación al Usuario**: Se envía una notificación al usuario a través de SQS una vez que el procesamiento se ha completado.



## Flujo de Trabajo

1. **Solicitud de Canción**: El usuario solicita reproducir una canción que no está almacenada en S3.
2. **Descarga desde YouTube**: El bot descarga la canción desde YouTube en formato .m4a.
3. **Subida a S3**: El archivo .m4a se carga en un bucket de S3.
4. **Activación de Lambda**: La subida del archivo activa una función Lambda.
5. **Recolección de Datos**: La Lambda obtiene información del archivo y actualiza el estado en DynamoDB a `PROCESSING`.
6. **Verificación de Estado**: Se comprueba en DynamoDB si la canción ya fue procesada. Si no, se envía la información a un servicio ECS para procesar el archivo con FFmpeg y DCA.
7. **Procesamiento en ECS**: ECS convierte el archivo .m4a a .dca y lo sube de nuevo a S3.
8. **Notificación al Usuario**: Una vez completado el procesamiento, se envía un mensaje de notificación al usuario a través de AWS SQS.

## Diagrama de Secuencia

![Diagrama de Secuencia](/docs/diagrama-job-ecs.png)


## Variables de Entorno

Algunos datos críticos se pasan de Lambda a ECS para su ejecución:

- `ACCESS_KEY`
- `CLUSTER_NAME`
- `SECRET_KEY`
- `SECURITY_GROUP`
- `TASK_DEFINITION`
- `TASK_EXECUTION_ARN`
- `TASK_ROLE_ARN`

## Notificación en Tiempo Real

Para notificar al usuario que la canción ha sido procesada, se utiliza AWS SQS. Una vez que el archivo es procesado y subido a S3, se envía un mensaje a una cola de SQS. El bot de música puede estar escuchando esa cola para reaccionar en tiempo real.

---

Esta feature permite que el bot de música funcione de manera eficiente, manteniendo la experiencia del usuario al minimizar tiempos de espera y optimizando el uso de recursos.