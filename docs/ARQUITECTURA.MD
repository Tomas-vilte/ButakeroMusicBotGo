### Configuración de EC2 para ButakeroMusicBotGo

🤖 **Descripción:**

El bot de Discord de ButakeroMusicBotGo está alojado en Amazon EC2, lo que proporciona un entorno virtual escalable y controlado para ejecutar aplicaciones y servicios.

🔧 **Componentes Principales:**

1. **VPC (Virtual Private Cloud)**: La instancia EC2 está dentro de una VPC que proporciona una red virtual aislada y segura para el bot. La VPC garantiza un control total sobre la configuración de red, incluyendo la asignación de direcciones IP, las tablas de enrutamiento y las puertas de enlace de Internet.

2. **Subnet**: La instancia EC2 se encuentra en una o más subredes de la VPC, distribuidas en distintas zonas de disponibilidad (AZs) para garantizar la alta disponibilidad y la tolerancia a fallos.

3. **Security Group**: Un grupo de seguridad está asociado a la instancia EC2 y controla el tráfico de red entrante y saliente. Se definen reglas de seguridad específicas para permitir o denegar el acceso a los puertos necesarios para el funcionamiento del bot.

🛠️ **Tecnologías Utilizadas:**

- **AWS EC2**: Se utiliza para alojar y ejecutar el bot de Discord en un entorno virtual escalable.
- **VPC, Subnet y Security Group**: Se emplean para configurar y asegurar el entorno de red de la instancia EC2.

🔄 **Flujo de Trabajo:**

1. La instancia EC2 se inicia con una configuración específica dentro de la VPC y las subredes definidas.
2. Se aplican las reglas de seguridad del grupo de seguridad asociado a la instancia para controlar el tráfico de red permitido.
3. El bot de Discord se ejecuta en la instancia EC2 y puede interactuar con otros servicios y recursos de AWS según sea necesario.

🔒 **Seguridad y Monitoreo:**

- Se aplican políticas de seguridad y reglas de seguridad específicas para restringir el acceso no autorizado a la instancia EC2.
- Se implementan medidas de monitoreo y registro para supervisar el rendimiento, la disponibilidad y la integridad de la instancia EC2 y sus recursos asociados.
---

### Configuración del Webhook de GitHub para ButakeroMusicBotGo

## Diagrama de Arquitectura

![Diagrama de Arquitectura](/images/Arquitectura%20Webhook.png)

**Descripción:**

ButakeroMusicBotGo utiliza un webhook de GitHub para recibir eventos de nuevas versiones o workflows. Estos eventos son procesados y utilizados para enviar notificaciones a los canales de Discord a través del bot.

🔧 **Componentes Principales:**

1. **API Gateway**: La API Gateway proporciona un punto de entrada seguro para los eventos de GitHub enviados a través del webhook. La ruta `/github-webhook` está configurada para recibir y enrutar estos eventos a una función Lambda.

2. **Funciones Lambda**: Una función Lambda se activa en respuesta a los eventos de GitHub recibidos a través de la API Gateway. Esta función procesa los eventos y los enruta a diferentes temas de SQS basados en el tipo de evento.

3. **Colas SQS**: Amazon SQS se utiliza para encolar los eventos de GitHub después de ser procesados por la función Lambda. Cada tipo de evento tiene su propio tema SQS para garantizar un procesamiento y enrutamiento eficientes.

🛠️ **Tecnologías Utilizadas:**

- **AWS API Gateway y Lambda**: Se utilizan para recibir, procesar y enrutar los eventos de GitHub a través del webhook.
- **Amazon SQS**: Se emplea para encolar y distribuir los eventos de GitHub a otras partes del sistema.

🔄 **Flujo de Trabajo:**

1. GitHub envía eventos de nuevas versiones o workflows a la API Gateway `/github-webhook`.
2. La función Lambda asociada a la API Gateway procesa estos eventos y los enruta a los temas SQS correspondientes.
3. Otras funciones Lambda están suscritas a los temas SQS y se activan cuando hay nuevos eventos en las colas. Estas funciones formatean los eventos y los envían a Discord a través del bot.

🔒 **Seguridad y Monitoreo:**

- Se aplican políticas de seguridad y reglas de seguridad específicas para proteger la API Gateway y las funciones Lambda.
- Se implementan medidas de monitoreo y registro para supervisar el rendimiento, la disponibilidad y la integridad de la API Gateway, las funciones Lambda y las colas SQS.

---

### Flujo de CI/CD para ButakeroMusicBotGo

🛠️ **Flujo de Desarrollo de Go**

## Diagrama de Secuencia CI/CD

![Diagrama de Secuencia CI/CD](/images/ci-cd%20d%20secuencia.png)

**Descripción:**

Este flujo de trabajo automatizado se encarga de compilar, probar y generar un informe de cobertura para el código fuente de ButakeroMusicBotGo cada vez que se realizan cambios en los archivos `.go`.

**Componentes Principales:**

- **GitHub Actions**: Se utiliza para ejecutar el flujo de CI/CD automáticamente en respuesta a eventos de push y pull request en la rama `master`.
- **Instalación de Dependencias**: Se instalan las dependencias necesarias para compilar el código, incluyendo herramientas como `golangci-lint`, `dca`, `yt-dlp`, entre otras.
- **Compilación y Pruebas**: Se compila el código, se ejecutan las pruebas unitarias y se genera un informe de cobertura para evaluar la calidad del código.
- **Subida de Informe de Cobertura**: Se sube el informe de cobertura generado para su revisión.

**🔁 Flujo de Entrega Continua de Go**

**Descripción:**

Este flujo de trabajo automatizado se encarga de desplegar la última versión del bot en un servidor EC2 después de que se haya completado satisfactoriamente el flujo de desarrollo.

**Componentes Principales:**

- **GitHub Actions**: Se utiliza para ejecutar el flujo de CI/CD automáticamente en respuesta a la finalización exitosa del flujo de desarrollo.
- **Conexión a AWS**: Se configuran las credenciales de AWS para interactuar con los servicios de EC2.
- **Construcción y Publicación de Imágenes Docker**: Se construye y publica una nueva imagen Docker con la última versión del bot en Docker Hub.
- **Despliegue a EC2**: Se utiliza SSH para conectarse al servidor EC2 y realizar el despliegue del bot actualizando el código y reiniciando los servicios necesarios.
---


### Herramientas de Monitoreo y Perfilado

📊 **Descripción:**

Además de la infraestructura y los flujos de CI/CD mencionados anteriormente, agregue herramientas de monitoreo y perfilado para supervisar y optimizar su rendimiento en tiempo real.

**Componentes Principales:**

- **Grafana**: Se utiliza como plataforma de visualización de datos para crear paneles y gráficos que muestran métricas importantes del bot, como el uso de CPU, la memoria, la latencia de red, etc.

- **Prometheus**: Se utiliza como sistema de monitorización y alerta para recopilar y almacenar métricas del bot. Se integra con Grafana para la visualización de datos en tiempo real.

- **Pyroscope**: Se utiliza para el perfilado de la aplicación, permitiendo identificar cuellos de botella y optimizar el rendimiento del código mediante el análisis de la CPU y el uso de memoria.

**🔍 Monitoreo y Optimización Continuos:**

Estas herramientas se utilizan de manera continua para monitorear el rendimiento del bot, identificar posibles problemas y realizar ajustes en la configuración y el código para garantizar un funcionamiento óptimo y una experiencia de usuario satisfactoria.

**Ejemplos Visuales:**

A continuación, se muestran algunas capturas de pantalla de los paneles en Grafana y Pyroscope:

#### Grafana:

![Panel de Grafana 1](/images/paneles.png)
1. **Panel 1: GC duration quantile**
   - **Descripción:** Muestra la duración de los ciclos de recolección de basura (GC) en diferentes percentiles.
   - **Explicación:** Indica cuánto tiempo tarda el recolector de basura en limpiar la memoria. Un aumento en la duración del GC puede señalar problemas de memoria o de rendimiento.

2. **Panel 2: Memory in Off-Heap**
   - **Descripción:** Muestra el uso de memoria fuera del montón (Off-Heap) en diferentes categorías.
   - **Explicación:** Muestra el uso de memoria en áreas de Go que no son parte del montón principal. Un aumento podría indicar un uso excesivo de recursos adicionales.

3. **Panel 3: Memory in Stack**
   - **Descripción:** Muestra el uso de memoria en la pila (Stack) en bytes.
   - **Explicación:** Indica cuánta memoria se está utilizando para almacenar datos locales de funciones, variables, etc. Un aumento podría indicar operaciones que requieren más memoria.

4. **Panel 4: Memory in Heap**
   - **Descripción:** Muestra el uso de memoria en el montón (Heap) en diferentes categorías.
   - **Explicación:** Muestra cómo se distribuye la memoria dentro del montón. Ayuda a identificar si se están asignando demasiados objetos o si hay memoria sin usar.

5. **Panel 5: Total Used Memory**
   - **Descripción:** Muestra la cantidad total de memoria usada por el bot.
   - **Explicación:** Indica la cantidad total de memoria que está utilizando el bot. Un aumento puede señalar problemas de memoria o de recursos.

6. **Panel 6: Número de Veces que se Usan los Comandos**
   - **Descripción:** Muestra el número de veces que se ejecuta un comando específico (PlaySong).
   - **Explicación:** Indica la frecuencia de uso de los comandos, proporcionando información sobre la interacción de los usuarios con el bot.

#### Pyroscope:

![Captura de Pyroscope](/images/profiling.png)
### Profiling

**Panel Superior:**
- **Descripción:** Muestra una gráfica del conteo total de objetos asignados en memoria a lo largo del tiempo. Los picos altos indican momentos donde se han asignado muchos objetos.
  
**Panel Inferior:**
- **Columna "Symbol":** Muestra el nombre de la función o paquete de código que ha asignado objetos en memoria.
- **Columna "Self":** Indica cuántos objetos ha asignado directamente la función o paquete.
- **Columna "Total":** Muestra el conteo total de objetos asignados por la función o paquete, incluyendo los objetos que se asignaron en las funciones que llama.

**Flame Graph:**
- **Descripción:** Visualización que muestra cómo se llama a las funciones en tu aplicación. Cada caja representa una función, y el tamaño es proporcional al número de objetos que asigna. Los colores ayudan a identificar diferentes paquetes de código.

**Información Extraída:**
- **Punto Caliente:** La función `github.com/Tomas-vite/GoMusicBot/internal/music/fetcher.(*YoutubeFetcher).GetDCAData.func1` parece ser la que más objetos asigna en memoria.
- **Llamadas:** El flame graph muestra que esta función llama a varias funciones, incluyendo `runtime/pprof.(*profileBuilder).flush` y `github.com/klauspost/compress/zip.(*Writer).Close`.
- **Posibles Problemas:** Funciones relacionadas con la compresión y la red también asignan muchos objetos, lo que podría indicar un uso excesivo o una fuga de memoria.