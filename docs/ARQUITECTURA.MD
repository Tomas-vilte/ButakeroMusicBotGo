> ⚠️ **Nota:** Esta arquitectura ya no está en uso actualmente. El sistema funciona, pero el bot ya no utiliza este flujo basado en webhooks de GitHub. La documentación se mantiene aquí solo como referencia histórica.

### Configuración del Webhook de GitHub para ButakeroMusicBotGo

## Diagrama de Arquitectura

![Diagrama de Arquitectura](/images/Arquitectura%20Webhook.png)

**Descripción:**

ButakeroMusicBotGo utiliza un webhook de GitHub para recibir eventos de nuevas versiones o workflows. Estos eventos son procesados y utilizados para enviar notificaciones a los canales de Discord a través del bot.

**Componentes Principales:**

1. **API Gateway**: La API Gateway proporciona un punto de entrada seguro para los eventos de GitHub enviados a través del webhook. La ruta `/github-webhook` está configurada para recibir y enrutar estos eventos a una función Lambda.

2. **Funciones Lambda**: Una función Lambda se activa en respuesta a los eventos de GitHub recibidos a través de la API Gateway. Esta función procesa los eventos y los enruta a diferentes temas de SQS basados en el tipo de evento.

3. **Colas SQS**: Amazon SQS se utiliza para encolar los eventos de GitHub después de ser procesados por la función Lambda. Cada tipo de evento tiene su propio tema SQS para garantizar un procesamiento y enrutamiento eficientes.

**Tecnologías Utilizadas:**

- **AWS API Gateway y Lambda**: Se utilizan para recibir, procesar y enrutar los eventos de GitHub a través del webhook.
- **Amazon SQS**: Se emplea para encolar y distribuir los eventos de GitHub a otras partes del sistema.

**Flujo de Trabajo:**

1. GitHub envía eventos de nuevas versiones o workflows a la API Gateway `/github-webhook`.
2. La función Lambda asociada a la API Gateway procesa estos eventos y los enruta a los temas SQS correspondientes.
3. Otras funciones Lambda están suscritas a los temas SQS y se activan cuando hay nuevos eventos en las colas. Estas funciones formatean los eventos y los envían a Discord a través del bot.

**Seguridad y Monitoreo:**

- Se aplican políticas de seguridad y reglas de seguridad específicas para proteger la API Gateway y las funciones Lambda.
- Se implementan medidas de monitoreo y registro para supervisar el rendimiento, la disponibilidad y la integridad de la API Gateway, las funciones Lambda y las colas SQS.
