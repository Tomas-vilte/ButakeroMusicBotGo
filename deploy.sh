#!/bin/bash
set -e

# ===========================
# Namespaces
# ===========================
echo "==========================="
echo "Creando namespaces..."
echo "==========================="
kubectl apply -f k8s/bases/kafka/namespace.yaml
kubectl apply -f k8s/bases/mongodb/namespace.yaml
kubectl apply -f k8s/bases/prometheus-operator/namespace.yaml
kubectl apply -f k8s/bases/backend/namespace.yaml

kubectl apply -f k8s/reflector.yaml

# ===========================
# RBAC
# ===========================
echo "==========================="
echo "Aplicando RBAC..."
echo "==========================="
kubectl apply -f k8s/bases/prometheus-operator/rbac/
kubectl apply -f k8s/bases/mongodb/rbac/

# ===========================
# CRDs
# ===========================
echo "==========================="
echo "Creando CRDs..."
echo "==========================="
kubectl create -f k8s/bases/prometheus-operator/crds
kubectl create -f k8s/bases/kafka/crd.yaml -n kafka
kubectl apply -f k8s/bases/mongodb/crd.yaml

echo "Esperando que los CRDs se establezcan..."
kubectl wait --for=condition=established --timeout=600s crd/kafkas.kafka.strimzi.io
kubectl wait --for=condition=established --timeout=600s crd/mongodbcommunity.mongodbcommunity.mongodb.com

# ===========================
# Cert-Manager
# ===========================
echo "==========================="
echo "Instalando Cert-Manager..."
echo "==========================="
kubectl apply -f k8s/bases/cert-manager/namespace.yaml
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-105 jetstack/cert-manager --namespace cert-manager --version v1.6.1 --values k8s/bases/cert-manager/helm-values.yaml

echo "Esperando que Cert-Manager esté listo..."
kubectl wait --for=condition=available --timeout=600s deployment/cert-105-cert-manager -n cert-manager
kubectl wait --for=condition=available --timeout=600s deployment/cert-105-cert-manager-cainjector -n cert-manager
kubectl wait --for=condition=available --timeout=600s deployment/cert-105-cert-manager-webhook -n cert-manager

# ===========================
# MongoDB
# ===========================
echo "==========================="
echo "Aplicando MongoDB..."
echo "==========================="
kubectl apply -f k8s/bases/mongodb/certificates
kubectl apply -f k8s/bases/mongodb/operator.yaml
kubectl apply -f k8s/bases/mongodb/internal
kubectl apply -f k8s/bases/mongodb/exporter

echo "Esperando que el operador de MongoDB esté listo..."
kubectl wait --for=condition=Ready --timeout=600s pod -l name=mongodb-kubernetes-operator -n mongodb
echo "Esperando a que se cree el pod de MongoDB..."
until kubectl get pods -n mongodb | grep -q my-mongodb-0; do
    echo "Esperando a que se cree el pod de MongoDB..."
    sleep 5
done

echo "Pod de MongoDB creado. Esperando a que esté listo..."
kubectl wait --for=condition=Ready --timeout=600s pod/my-mongodb-0 -n mongodb


# ===========================
# Kafka
# ===========================
echo "==========================="
echo "Aplicando Kafka..."
echo "==========================="
kubectl apply -f k8s/bases/kafka/certificates
kubectl apply -f k8s/bases/kafka/internal/kafka.yaml -n kafka
kubectl apply -f k8s/bases/kafka/topic.yaml
kubectl apply -f k8s/bases/kafka/user.yaml
kubectl apply -f k8s/bases/kafka/strmizi-pod-monitor.yaml

echo "Esperando que Kafka esté listo..."
kubectl wait --for=condition=Ready --timeout=600s pod -l name=strimzi-cluster-operator -n kafka

echo "Esperando a que se creen los pods de Kafka..."
until kubectl get pods -n kafka | grep -q kafka; do
    echo "Esperando a que se creen los pods de Kafka..."
    sleep 5
done

echo "Pods de Kafka creados. Esperando a que estén listos..."
kubectl wait --for=condition=Ready --timeout=600s pod -l app.kubernetes.io/name=kafka -n kafka
kubectl wait --for=condition=Ready --timeout=600s pod -l app.kubernetes.io/name=zookeeper -n kafka

# ===========================
# Prometheus & Grafana
# ===========================
echo "==========================="
echo "Aplicando Prometheus y Grafana..."
echo "==========================="
kubectl apply -f k8s/bases/prometheus-operator/deployment
kubectl apply -f k8s/bases/prometheus/
kubectl apply -R -f k8s/bases/grafana

# ===========================
# Otros recursos
# ===========================
echo "==========================="
echo "Aplicando otros recursos..."
echo "==========================="
kubectl apply -f k8s/bases/cadvisor

# ===========================
# Backend
# ===========================
echo "==========================="
echo "Aplicando Backend..."
echo "==========================="
kubectl apply -f k8s/bases/backend/

echo "Esperando que los pods de Backend estén listos..."
kubectl wait --for=condition=Ready --timeout=600s pod -l app=backend-processing-audio -n backend

# ===========================
# Verificación Final
# ===========================
echo "==========================="
echo "Verificación final de los pods y recursos desplegados..."
echo "==========================="
kubectl get pods --all-namespaces
kubectl get crds

echo "==========================="
echo "Despliegue completado exitosamente."
echo "==========================="
