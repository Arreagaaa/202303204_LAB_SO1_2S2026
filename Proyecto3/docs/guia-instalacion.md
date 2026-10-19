# Guia de instalacion y despliegue

**Autor:** Christian Javier Rivas Arreaga · **Carné:** 202303204.

## 1. Objetivo

Esta guia explica como preparar el ambiente, validar el codigo, construir las imagenes, publicarlas en Zot y desplegar los componentes en Kubernetes.

Los comandos de compilacion y validacion se ejecutan desde WSL2 cuando el equipo principal es Windows. Los comandos de GCP y Kubernetes pueden ejecutarse desde WSL2 o desde una distribucion Linux con acceso a las credenciales correspondientes.

## 2. Requisitos

### 2.1. Herramientas locales

Para la prueba local de extremo a extremo basta con Docker Engine y Docker Compose dentro de WSL2; los Dockerfiles compilan Go y Rust. Las otras herramientas se necesitan para desarrollo, validaciones individuales o despliegue:

- Git
- Go 1.22 o una version compatible
- Rust y Cargo 1.99 (version fijada en el Dockerfile)
- Docker Engine con el plugin Compose, o Docker Desktop con integracion para WSL2
- kubectl
- Helm
- Locust, solo si se ejecuta fuera del contenedor de Compose
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

Los archivos protobuf generados ya estan versionados. Si se modifica `proto/report.proto`, regenerarlos con `bash scripts/generate-proto.sh` antes de compilar.

Compilar cada servicio:

```bash
go build ./go-ingest
go build ./go-ingest-client
go build ./go-writer
go build ./go-writer-publisher
go build ./consumer
go build ./metrics
cargo build --manifest-path api-rust/Cargo.toml
```

Ejecutar la validacion agrupada:

```bash
bash scripts/test-wsl.sh
```

El script no crea recursos en GCP. Solo revisa y compila los componentes locales.

## 4. Dependencias locales

Desde la raiz de `Proyecto3`, iniciar todo el flujo local:

```bash
docker compose -f compose.local.yaml up -d --build
```

Comprobar los contenedores y probar la API:

```bash
docker compose -f compose.local.yaml ps
curl http://localhost:8080/health
```

La API queda en `http://localhost:8080`, RabbitMQ Management en `http://localhost:15672` y Grafana en `http://localhost:3300`. Usuario local: `mumnk8s` para RabbitMQ y `admin` para Grafana; ambas contrasenas locales son `x`. Esas credenciales son solo para la prueba local.

Ejecutar la prueba de carga de un minuto y detener el ambiente:

```bash
docker compose -f compose.local.yaml --profile load run --rm locust
docker compose -f compose.local.yaml down
```

### 4.1. Kubernetes local con kind

En WSL se probo un cluster `kind` llamado `mumnk8s-test`. Las herramientas `kind`, `kubectl` y `helm` se instalaron en `~/.local/bin`. Construir primero las ocho imagenes de aplicacion con las etiquetas `mumnk8s/<servicio>:local` y ejecutar:

```bash
kind create cluster --name mumnk8s-test --wait 120s
bash scripts/load-kind-images.sh mumnk8s-test
kubectl --context kind-mumnk8s-test apply -f k8s/00-namespace.yaml -f k8s/07-crds.yaml
kubectl --context kind-mumnk8s-test apply -f k8s/local/rabbitmq-secret.yaml -f k8s/local/grafana-secret.yaml
sed 's#zot.example/mumnk8s/rabbitmq:3.13-management#rabbitmq:3.13-management#' k8s/01-rabbitmq.yaml | kubectl --context kind-mumnk8s-test apply -f -
kubectl --context kind-mumnk8s-test apply -f k8s/02-services.yaml -f k8s/05-consumer-valkey.yaml -f k8s/09-operator.yaml
sed 's#zot.example/mumnk8s/prometheus:v2.54.1#prom/prometheus:v2.54.1#' k8s/10-metrics.yaml | kubectl --context kind-mumnk8s-test apply -f -
kubectl --context kind-mumnk8s-test apply -f k8s/local/custom-resources.yaml
kubectl --context kind-mumnk8s-test get pods,pvc -n mumnk8s
```

`scripts/load-kind-images.sh` carga las imagenes propias con los nombres esperados por los manifiestos, sin publicar en Zot. RabbitMQ, Valkey, Prometheus y Grafana se obtienen de sus repositorios publicos solo en esta prueba local. Esperar a que todos los pods esten `Ready`; el publicador y el consumer pueden reiniciarse mientras RabbitMQ termina de arrancar.

Para probar Gateway API en kind se uso Envoy Gateway. La clase `gke-l7-global-external-managed` del manifiesto principal es exclusiva de GKE; localmente se reemplaza solo al aplicar:

```bash
helm --kube-context kind-mumnk8s-test install eg oci://docker.io/envoyproxy/gateway-helm --version v1.9.2 -n envoy-gateway-system --create-namespace
kubectl --context kind-mumnk8s-test apply -f k8s/local/gateway-class.yaml
sed 's/gke-l7-global-external-managed/envoy-local/' k8s/04-gateway.yaml | kubectl --context kind-mumnk8s-test apply -f -
kubectl --context kind-mumnk8s-test get gateway,httproute -n mumnk8s
```

La ruta puede probarse con `kubectl port-forward` al Service `envoy-mumnk8s-mumnk8s-gateway-...` en el namespace `envoy-gateway-system`. En kind, el Service LoadBalancer no obtiene IP externa; esto deja `Gateway Programmed=False` aunque la ruta HTTP funciona. Para HPA se instalo Metrics Server v0.8.0 y se anadio el argumento `--kubelet-insecure-tls`, limitado al cluster de prueba. Luego se aplico `k8s/03-hpa.yaml` y se comprobo `kubectl get hpa -n mumnk8s` durante Locust.

Eliminar el cluster cuando ya no se necesite con `kind delete cluster --name mumnk8s-test`; esa accion borra tambien sus datos locales.

## 5. Construir las imagenes

El contexto de Docker debe ser la carpeta `Proyecto3`, porque los Dockerfiles de Go necesitan acceder al modulo raiz y al contrato protobuf.

Definir el nombre del registry:

```bash
export ZOT=registry.example.com
sed -i "s#zot.example#$ZOT#g" \
  operator/helm-charts/valkeyinstance/values.yaml \
  operator/helm-charts/grafanainstance/values.yaml
```

Construir las imagenes:

```bash
docker build -f api-rust/Dockerfile -t "$ZOT/mumnk8s/api-rust:dev" .
docker build -f go-ingest/Dockerfile -t "$ZOT/mumnk8s/go-ingest:dev" .
docker build -f go-ingest-client/Dockerfile -t "$ZOT/mumnk8s/go-ingest-client:dev" .
docker build -f go-writer/Dockerfile -t "$ZOT/mumnk8s/go-writer:dev" .
docker build -f go-writer-publisher/Dockerfile -t "$ZOT/mumnk8s/go-writer-publisher:dev" .
docker build -f consumer/Dockerfile -t "$ZOT/mumnk8s/consumer:dev" .
docker build -f metrics/Dockerfile -t "$ZOT/mumnk8s/metrics:dev" .
docker build -f operator/Dockerfile -t "$ZOT/mumnk8s/operator:dev" operator
```

Verificar que las imagenes existan:

```bash
docker image ls "$ZOT/mumnk8s/*"
```

Publicarlas en Zot:

```bash
docker push "$ZOT/mumnk8s/api-rust:dev"
docker push "$ZOT/mumnk8s/go-ingest:dev"
docker push "$ZOT/mumnk8s/go-ingest-client:dev"
docker push "$ZOT/mumnk8s/go-writer:dev"
docker push "$ZOT/mumnk8s/go-writer-publisher:dev"
docker push "$ZOT/mumnk8s/consumer:dev"
docker push "$ZOT/mumnk8s/metrics:dev"
docker push "$ZOT/mumnk8s/operator:dev"
```

Publicar tambien las imagenes de infraestructura, porque el cluster debe descargarlas desde Zot:

```bash
for image in \
  'rabbitmq:3.13-management|rabbitmq:3.13-management' \
  'valkey/valkey:8|valkey:8' \
  'grafana/grafana:11.2.0|grafana:11.2.0' \
  'prom/prometheus:v2.54.1|prometheus:v2.54.1'; do
  source_image="${image%%|*}"
  target_image="$ZOT/mumnk8s/${image##*|}"
  docker pull "$source_image"
  docker tag "$source_image" "$target_image"
  docker push "$target_image"
done
```

Antes de publicar, validar que el certificado HTTPS de Zot sea confiable para Docker y para el runtime de nodos de GKE.

## 6. Configurar las imagenes de Kubernetes

Los manifiestos contienen el dominio de ejemplo `zot.example`. Antes de aplicarlos, reemplazarlo por el dominio real:

```bash
sed -i "s#zot.example#$ZOT#g" \
  k8s/01-rabbitmq.yaml k8s/02-services.yaml k8s/05-consumer-valkey.yaml \
  k8s/09-operator.yaml k8s/10-metrics.yaml
```

Los `values.yaml` del operador se ajustan antes de construir su imagen porque se copian dentro de ella. Los manifiestos de produccion no incluyen contrasenas: `rabbitmq-credentials` y `grafana-admin` deben crearse en el namespace `mumnk8s` con claves reales, fuera de Git, antes de instalar los servicios. Los secretos en `k8s/local/` usan credenciales desechables exclusivamente para kind.

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
# Crear rabbitmq-credentials (username, password, uri) y grafana-admin
# (admin-password) desde un gestor de secretos o archivos locales protegidos.
kubectl apply -f k8s/01-rabbitmq.yaml
kubectl apply -f k8s/02-services.yaml
kubectl apply -f k8s/03-hpa.yaml
kubectl apply -f k8s/04-gateway.yaml
kubectl apply -f k8s/05-consumer-valkey.yaml
kubectl apply -f k8s/10-metrics.yaml
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

El procedimiento detallado esta en [operator/README.md](../operator/README.md). Aplicar los CRDs, el operador y luego los recursos:

```bash
kubectl apply -f k8s/07-crds.yaml
kubectl apply -f k8s/09-operator.yaml
kubectl rollout status deployment/mumnk8s-operator -n mumnk8s
kubectl apply -f k8s/06-custom-resources.yaml
```

Comprobar los recursos:

```bash
kubectl get valkeyinstances,grafanainstances -n mumnk8s
kubectl get statefulsets,deployments,svc,pvc -n mumnk8s
```

Para abrir Grafana desde el equipo con acceso a `kubectl`:

```bash
kubectl port-forward -n mumnk8s svc/grafana-dashboard 3000:3000
```

Abrir `http://127.0.0.1:3000` y comprobar el dashboard `M.U.M.N.K8s - Reportes militares` con datos de Locust. Grafana consulta a Prometheus, que a su vez lee las métricas calculadas desde Valkey por el exporter.

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
kubectl exec -n mumnk8s statefulset/valkey-reports -- valkey-cli keys 'reports:*'
```
