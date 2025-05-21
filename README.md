# ButakeroMusicBotGo 

**ButakeroMusicBotGo** es un bot de Discord desarrollado en GoLang que permite reproducir música en servidores de Discord. Este repositorio contiene el código fuente del bot junto con instrucciones para instalarlo y usarlo. Actualmente tiene compatibilidad con Youtube, pero mas adelante voy a agregar otras plataformas :D

## 🏗️ Arquitectura del Bot

¿Querés chusmear la arquitectura del bot en producción? Hacé clic [aca](/images/ARQUITECTURA.MD) para ver todos los detalles sobre cómo está construido y desplegado ButakeroMusicBotGo.

## 🤖 Invitación al Bot

Aca tenes la invitacion para probar ButakeroMusicBotGo en tu servidor de Discord, usa este enlance para invitarlo:
[Invitacion del bot a tu server](https://discord.com/oauth2/authorize?client_id=987850036866084974)

## 🚀 Instalación

### ⚙️ Configuración del bot en el portal de desarrolladores de Discord

1. Ve a [Discord Developer Portal](https://discord.com/developers/applications).

2. Hacé clic en "New Application" y dale un nombre a tu aplicación.

3. En la pestaña de "Installation", marcá las casillas "User install" y "Guild install".

4. En la pestaña activa las siguientes opciones:
   - **PUBLIC BOT**: Permití que el bot sea agregado por cualquiera. Cuando no está marcado, solo vos podés agregar este bot a servidores.
   - **PRESENCE INTENT**: Necesario para que tu bot reciba eventos de actualización de presencia.
     - **Nota**: Una vez que tu bot llegue a 100 o más servidores, esto requerirá verificación y aprobación. Leé más [aquí](https://support-dev.discord.com/hc/en-us/articles/6205754771351-How-do-I-get-Privileged-Intents-for-my-bot).
   - **SERVER MEMBERS INTENT**: Necesario para que tu bot reciba eventos listados bajo GUILD_MEMBERS.
     - **Nota**: Una vez que tu bot llegue a 100 o más servidores, esto requerirá verificación y aprobación. Leé más [aquí](https://support-dev.discord.com/hc/en-us/articles/6205754771351-How-do-I-get-Privileged-Intents-for-my-bot).
   - **MESSAGE CONTENT INTENT**: Necesario para que tu bot reciba el contenido de los mensajes en la mayoría de los mensajes.
     - **Nota**: Una vez que tu bot llegue a 100 o más servidores, esto requerirá verificación y aprobación. Leé más [aquí](https://support-dev.discord.com/hc/en-us/articles/6205754771351-How-do-I-get-Privileged-Intents-for-my-bot).

5. Copiá el `DISCORDTOKEN` de la sección de Bot y guardalo, lo necesitarás para configurar el archivo `.env`.

### 🐳 Ejecución con Docker Compose

Para ejecutar **ButakeroMusicBotGo** utilizando Docker Compose, seguí estos pasos:

1. Asegurate de tener Docker y Docker Compose instalados en tu sistema. Podés encontrar instrucciones de instalación en [Docker](https://docs.docker.com/get-docker/) y [Docker Compose](https://docs.docker.com/compose/install/).

2. Cloná este repositorio a tu máquina local:

    ```
    git clone git@github.com:Tomas-vilte/ButakeroMusicBotGo.git
    ```

3. Navegá hasta el directorio del repositorio clonado:

    ```
    cd ButakeroMusicBotGo
    ```

4. Creá un archivo `.env` utilizando el archivo de ejemplo proporcionado `.env.example`. Este archivo debería contener las siguientes variables:
    - `DISCORDTOKEN`: El token del bot que obtuviste en el portal de desarrolladores de Discord.
    - `COMMANDPREFIX`: El prefijo de comando que desees utilizar (por ejemplo, `/bot`).

5. Ejecutá el siguiente comando para construir los contenedores Docker:

    ```
    docker-compose --env-file .env -f local-docker-compose.yml build
    ```

6. Una vez que se haya completado la construcción, podés levantar todos los servicios necesarios (bot de Discord, servicios de monitoreo etc) con el siguiente comando:

    ```
    docker-compose --env-file .env -f local-docker-compose.yml up
    ```

## 🎧 Uso

Una vez que el bot esté en funcionamiento, podés interactuar con él en tu servidor de Discord. Acá tenés algunos comandos básicos que podés usar:

- `/seso play <nombre de la canción>`: Reproduce una canción en el canal de voz actual.
- `/seso stop`: Detiene la reproducción actual y desconecta el bot del canal de voz.
- `/seso list`: Muestra la lista de reproducción actual.
- `/seso skip`: Salta a la siguiente canción en la lista de reproducción.
- `/seso remove <número>`: Elimina una canción específica de la lista de reproducción.
- `/seso playing`: Muestra información sobre la canción que se está reproduciendo actualmente.

## 🤝 Contribuciones

¡Se agradecen las contribuciones! Si querés contribuir en el proyecto, seguí estos pasos:

1. Hacete un fork de este repositorio.
2. Realizá tus cambios en una nueva rama.
3. Envía un PR con una descripción clara de tus cambios.

---
