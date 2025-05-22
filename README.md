# ButakeroMusicBotGo

**ButakeroMusicBotGo** es un bot de Discord que hice en Go para que puedas escuchar música en tu servidor de Discord. Este repo tiene el código fuente del bot y las instrucciones para instalarlo y ponerlo a funcionar. Ahora mismo funciona con YouTube, pero en el futuro tengo pensado agregar otras plataformas. :D

## 🏗️ Arquitectura del Bot

¿Te pinta chusmear la arquitectura del bot en producción? Podés ver todos los detalles sobre cómo está armado y desplegado ButakeroMusicBotGo [aca](/images/ARQUITECTURA.MD).

## 🤖 Invitación al Bot

Si querés probar el bot en tu servidor de Discord, acá te dejo la invitación para que lo invites:

[Invitación del bot a tu server](https://discord.com/oauth2/authorize?client_id=987850036866084974)

## 🚀 Instalación

### ⚙️ Configuración del bot en el portal de desarrolladores de Discord

1. Primero, andá al [Discord Developer Portal](https://discord.com/developers/applications).

2. Hacé clic en "New Application" y poné el nombre que quieras para tu aplicación.

3. En la pestaña de "Installation", activá las casillas "User install" y "Guild install".

4. En la pestaña, activá estas opciones:
    - **PUBLIC BOT**: Para que cualquiera pueda agregar el bot a otros servidores. Si no está activado, solo vos lo podés agregar a tu servidor.
    - **PRESENCE INTENT**: Necesario para que tu bot pueda recibir eventos de actualización de presencia.
        - **Nota**: Si el bot llega a 100 o más servidores, vas a tener que pedir una verificación. Leé más [aquí](https://support-dev.discord.com/hc/en-us/articles/6205754771351-How-do-I-get-Privileged-Intents-for-my-bot).
    - **SERVER MEMBERS INTENT**: Necesario para que tu bot reciba eventos listados bajo GUILD_MEMBERS.
        - **Nota**: Lo mismo, si el bot llega a 100 o más servidores, necesitarás la verificación.
    - **MESSAGE CONTENT INTENT**: Necesario para que el bot lea el contenido de los mensajes en la mayoría de los casos.
        - **Nota**: También necesitarás la verificación si el bot llega a 100 o más servidores.

5. Copiate el `DISCORDTOKEN` de la sección de Bot y guardalo. Lo vas a necesitar para configurar el archivo `.env`.

### 🐳 Ejecución con Docker Compose: Orquestando ButakeroMusicBotGo

Para poner a andar el bot y sus microservicios, vamos a usar Docker Compose. Esta configuración está pensada para facilitar el desarrollo y las pruebas locales, encapsulando todas las dependencias necesarias.

1. Primero, asegurate de tener Docker y Docker Compose instalados en tu máquina. Si no los tenés, podés seguir estas guías: [Docker Engine](https://docs.docker.com/get-docker/) y [Docker Compose](https://docs.docker.com/compose/install/).

2. Cloná este repo a tu entorno local:

    ```bash
    git clone git@github.com:Tomas-vilte/ButakeroMusicBotGo.git
    ```

3. Entrá en el directorio del repositorio:

    ```bash
    cd ButakeroMusicBotGo
    ```

4. **Configuración de Variables de Entorno**:  
   Creá un archivo `.env` en la raíz del repositorio, basándote en el `.env.example` que viene con el repo. Este archivo tiene que tener estas variables esenciales. También podés exportarlas en tu terminal o configurarlas como variables de entorno del sistema:

    * `DISCORDTOKEN`: El token de autenticación para el bot de Discord. Este es esencial para que el bot funcione.
    * `COMMANDPREFIX`: El prefijo configurable para los comandos del bot (ej: `/seso`).
    * `YOUTUBE_API_KEY`: Tu clave de API de YouTube. **Es muy importante** para que el microservicio `audio_processor` pueda buscar y procesar contenido de YouTube.

   **Nota Avanzada (Opcional)**:  
   Si tenés restricciones de descarga o autenticación con YouTube (por ejemplo, contenido bloqueado por región o edad), podés crear un archivo `yt-cookies.txt` en la raíz del repositorio. Este archivo será montado en el contenedor `audio_processor` para su uso.

5. **Lanzamiento de Servicios (Up)**:  
   Una vez que Docker construya las imágenes, podés iniciar todos los servicios con:

    ```bash
    docker-compose --env-file .env up
    ```

   Esto levantará la siguiente arquitectura de microservicios, pensada para ser escalable y robusta:

    - 🐘 `zookeeper`: Esencial para coordinar los servicios de Kafka.
    - ⚙️ `kafka`: El broker que maneja los mensajes entre los microservicios.
    - 💾 `mongodb`: Base de datos NoSQL para guardar metadatos de canciones y más.
    - 🎶 `audio_processor`: Microservicio encargado de descargar y procesar audio.
    - 🤖 `butakero_bot`: El corazón del bot, que interactúa con Discord y gestiona la cola de reproducción.

---

**Volúmenes Persistentes para la Durabilidad de Datos:**

- `mongo_data`: Asegura que no se pierdan los datos de MongoDB al reiniciar el contenedor.
- `audio_files`: Comparte los archivos de audio procesados entre el `audio_processor` y el `butakero_bot`.

**Red de Docker (`test-application`):**

- `test-application`: Red personalizada para que todos los servicios se comuniquen entre sí de manera sencilla.

## 🎧 Uso

Una vez que el bot esté andando, podés interactuar con él en tu servidor de Discord usando estos comandos básicos:

- `/seso play <nombre de la canción>`: Reproduce una canción en el canal de voz.
- `/seso stop`: Detiene la reproducción y desconecta al bot.
- `/seso list`: Muestra la lista de reproducción.
- `/seso skip`: Salta a la siguiente canción.
- `/seso remove <número>`: Elimina una canción de la lista.
- `/seso playing`: Muestra la canción que está sonando.

## 🤝 Contribuciones

¡Todo aporte es bienvenido! Si querés contribuir, seguí estos pasos:

1. Hacete un fork de este repo.
2. Hacé tus cambios en una nueva rama.
3. Mandame un PR con una descripción clara de lo que hiciste.

---
