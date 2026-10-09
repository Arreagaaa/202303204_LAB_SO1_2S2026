# Guia de instalacion y despliegue

## 1. Objetivo

Esta guia explica como preparar el ambiente, validar el codigo, construir las imagenes, publicarlas en Zot y desplegar los componentes en Kubernetes.

Los comandos de compilacion y validacion se ejecutan desde WSL2 cuando el equipo principal es Windows. Los comandos de GCP y Kubernetes pueden ejecutarse desde WSL2 o desde una distribucion Linux con acceso a las credenciales correspondientes.

## 2. Requisitos

### 2.1. Herramientas locales

Instalar las siguientes herramientas dentro de WSL2:

- Git
- Go 1.22 o una version compatible
- Rust y Cargo
- Docker o Docker Desktop con integracion para WSL2
- kubectl
- Helm
- Locust
- protoc y los plugins de Go si se regeneran archivos protobuf
- kubeconform, opcional para validar manifiestos
- oras, necesario para el artifact OCI

Verificar las instalaciones:

```bash
go version
rustc --version
cargo --version
docker --version
kubectl version --client
helm version
locust --version
```

### 2.2. Infraestructura de destino

Para el despliegue final se necesita:

1. Un proyecto de Google Cloud con facturacion habilitada.
2. Un cluster GKE.
3. Una VM N1 fuera del cluster para Zot.
4. Un registry Zot accesible por HTTPS.
5. Permisos para crear recursos en GKE y publicar imagenes.
6. Una GatewayClass compatible con Gateway API en GKE.

Los nombres de proyecto, cluster, VM, registry y dominio deben definirse antes de ejecutar los comandos de despliegue.

## 3. Preparar el codigo

Desde WSL2, ubicarse en la raiz del proyecto:

```bash
cd /mnt/d/2S2026/SOPES1/LAB/202303204_LAB_SO1_2S2026/Proyecto3
```

Sincronizar los modulos de Go:

```bash
go work sync
```

Compilar cada servicio:

```bash
go build ./go-ingest
go build ./go-writer
go build ./consumer
cargo build --manifest-path api-rust/Cargo.toml
```

Ejecutar la validacion agrupada:

```bash
bash scripts/test-wsl.sh
```

El script no crea recursos en GCP. Solo revisa y compila los componentes locales.

## 4. Dependencias locales

Para una prueba de integracion local se pueden iniciar RabbitMQ y Valkey:

```bash
docker run -d --name mumnk8s-rabbitmq \
  -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management

docker run -d --name mumnk8s-valkey \
  -p 6379:6379 valkey/valkey:8
```

Comprobar que ambos contenedores esten activos:

```bash
docker ps --filter name=mumnk8s-rabbitmq --filter name=mumnk8s-valkey
```

Para limpiar el ambiente local:

```bash
docker rm -f mumnk8s-rabbitmq mumnk8s-valkey
```

## 5. Construir las imagenes

El contexto de Docker debe ser la carpeta `Proyecto3`, porque los Dockerfiles de Go necesitan acceder al modulo raiz y al contrato protobuf.

Definir el nombre del registry:

```bash
export ZOT=registry.example.com
```

Construir las imagenes:

```bash
docker build -f api-rust/Dockerfile -t "$ZOT/mumnk8s/api-rust:dev" .
docker build -f go-ingest/Dockerfile -t "$ZOT/mumnk8s/go-ingest:dev" .
docker build -f go-writer/Dockerfile -t "$ZOT/mumnk8s/go-writer:dev" .
docker build -f consumer/Dockerfile -t "$ZOT/mumnk8s/consumer:dev" .
```

Verificar que las imagenes existan:

```bash
docker image ls "$ZOT/mumnk8s/*"
```

Publicarlas en Zot:

```bash
docker push "$ZOT/mumnk8s/api-rust:dev"
docker push "$ZOT/mumnk8s/go-ingest:dev"
docker push "$ZOT/mumnk8s/go-writer:dev"
docker push "$ZOT/mumnk8s/consumer:dev"
```

Antes de publicar, validar que el certificado HTTPS de Zot sea confiable para Docker y para el runtime de nodos de GKE.

## 6. Configurar las imagenes de Kubernetes

Los manifiestos contienen el dominio de ejemplo `zot.example`. Antes de aplicarlos, reemplazarlo por el dominio real:

```bash
sed -i "s#zot.example#$ZOT#g" k8s/02-services.yaml k8s/05-consumer-valkey.yaml
```

Revisar los cambios antes de continuar:

```bash
grep -R "image:" k8s
```

## 7. Desplegar el flujo principal

Configurar primero el contexto de kubectl:

```bash
gcloud container clusters get-credentials NOMBRE_DEL_CLUSTER \
  --region REGION \
  --project ID_DEL_PROYECTO
```

Aplicar los manifiestos en orden:

```bash
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-rabbitmq.yaml
kubectl apply -f k8s/02-services.yaml
kubectl apply -f k8s/03-hpa.yaml
kubectl apply -f k8s/04-gateway.yaml
kubectl apply -f k8s/05-consumer-valkey.yaml
```

Revisar el estado:

```bash
kubectl get pods -n mumnk8s
kubectl get svc -n mumnk8s
kubectl get pvc -n mumnk8s
kubectl get hpa -n mumnk8s
kubectl get gateway -n mumnk8s
```

No continuar con la prueba de carga hasta que los pods principales esten `Ready` y el Gateway tenga una direccion asignada.

## 8. Instalar el Operator y sus recursos

El procedimiento detallado esta en [operator/README.md](../operator/README.md). Una vez instalado el proyecto generado con Operator SDK, aplicar primero las definiciones de recursos:

```bash
kubectl apply -f k8s/07-crds.yaml
kubectl apply -f k8s/06-custom-resources.yaml
```

Comprobar los recursos:

```bash
kubectl get valkeyinstances,grafanainstances -n mumnk8s
kubectl get statefulsets,deployments,svc,pvc -n mumnk8s
```

## 9. Instalar Falco

La instalacion y la prueba de la regla personalizada estan documentadas en [falco.md](falco.md):

```bash
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update
helm upgrade --install falco falcosecurity/falco \
  --namespace falco --create-namespace \
  --values k8s/falco-values.yaml
```

## 10. Ejecutar Locust

Obtener la direccion del Gateway:

```bash
kubectl get gateway mumnk8s-gateway -n mumnk8s
```

Ejecutar la prueba:

```bash
locust -f locust/locustfile.py --host http://IP_DEL_GATEWAY
```

Registrar para la evidencia:

- Numero de usuarios.
- Tasa de solicitudes.
- Porcentaje de errores.
- Latencia promedio y percentiles.
- Replicas observadas durante la prueba.
- Estado del HPA antes, durante y despues.

## 11. Diagnostico rapido

### Pods que no inician

```bash
kubectl describe pod NOMBRE_DEL_POD -n mumnk8s
kubectl logs NOMBRE_DEL_POD -n mumnk8s
```

### API sin respuesta

```bash
kubectl get endpoints api-rust -n mumnk8s
kubectl port-forward svc/api-rust 8080:80 -n mumnk8s
curl -i http://127.0.0.1:8080/health
```

### HPA sin metricas

```bash
kubectl get hpa -n mumnk8s
kubectl top pods -n mumnk8s
```

Si `kubectl top` no devuelve datos, revisar Metrics Server antes de evaluar el escalamiento.

### Mensajes no procesados

```bash
kubectl logs deploy/go-writer -n mumnk8s
kubectl logs deploy/consumer -n mumnk8s
kubectl exec -n mumnk8s statefulset/valkey -- valkey-cli keys 'reports:*'
```
