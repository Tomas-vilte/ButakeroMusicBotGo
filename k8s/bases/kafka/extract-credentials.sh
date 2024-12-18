#!/bin/bash

# Configuración
NAMESPACE=kafka
USER_SECRET=my-kafka-user
CA_SECRET=kafka-server-cert
TRUSTSTORE_PASSWORD=123456
USER_ALIAS=my-user

# Archivos generados
TRUSTSTORE_FILE=truststore.jks
USER_P12_FILE=user.p12
TEMP_CA=ca.crt
TEMP_USER_CERT=user.crt
TEMP_USER_KEY=user.key

# Función para verificar herramientas necesarias
check_command() {
    if ! command -v "$1" &> /dev/null; then
        echo "Error: La herramienta '$1' no está instalada o no está en el PATH."
        exit 1
    fi
}

# Verificar herramientas necesarias
for cmd in kubectl base64 keytool openssl; do
    check_command "$cmd"
done

# Limpieza inicial
rm -f "$TRUSTSTORE_FILE" "$USER_P12_FILE" "$TEMP_CA" "$TEMP_USER_CERT" "$TEMP_USER_KEY"

# Extraer el certificado de la CA
echo "Extrayendo el certificado de la CA..."
if ! kubectl get secret "$CA_SECRET" -n "$NAMESPACE" -o jsonpath='{.data.ca\.crt}' | base64 -d > "$TEMP_CA"; then
    echo "Error: No se pudo extraer el certificado de la CA."
    exit 1
fi
echo "Certificado de la CA extraído exitosamente."

# Extraer el certificado del usuario
echo "Extrayendo el certificado del usuario..."
if ! kubectl get secret "$USER_SECRET" -n "$NAMESPACE" -o jsonpath='{.data.user\.crt}' | base64 -d > "$TEMP_USER_CERT"; then
    echo "Error: No se pudo extraer el certificado del usuario."
    exit 1
fi

# Extraer la clave privada del usuario
echo "Extrayendo la clave privada del usuario..."
if ! kubectl get secret "$USER_SECRET" -n "$NAMESPACE" -o jsonpath='{.data.user\.key}' | base64 -d > "$TEMP_USER_KEY"; then
    echo "Error: No se pudo extraer la clave privada del usuario."
    exit 1
fi
echo "Certificado y clave privada del usuario extraídos exitosamente."

# Importar el certificado de la CA en el truststore
echo "Importando el certificado de la CA en el truststore..."
if ! echo "yes" | keytool -import -trustcacerts -file "$TEMP_CA" -keystore "$TRUSTSTORE_FILE" -storepass "$TRUSTSTORE_PASSWORD" -noprompt; then
    echo "Error: No se pudo importar el certificado de la CA en el truststore."
    exit 1
fi
echo "Certificado de la CA importado exitosamente en el truststore."

# Crear archivo PKCS12 con el certificado y la clave privada del usuario
echo "Generando archivo PKCS12 para el usuario..."
if ! RANDFILE=/tmp/.rnd openssl pkcs12 -export -in "$TEMP_USER_CERT" -inkey "$TEMP_USER_KEY" -name "$USER_ALIAS" -password pass:"$TRUSTSTORE_PASSWORD" -out "$USER_P12_FILE"; then
    echo "Error: No se pudo generar el archivo PKCS12 del usuario."
    exit 1
fi
echo "Archivo PKCS12 del usuario generado exitosamente."

# Limpieza de archivos temporales
rm -f "$TEMP_CA" "$TEMP_USER_CERT" "$TEMP_USER_KEY"
echo "Archivos temporales eliminados."

echo "Extracción de credenciales completada exitosamente. Archivos generados:"
echo " - Truststore: $TRUSTSTORE_FILE"
echo " - PKCS12: $USER_P12_FILE"
