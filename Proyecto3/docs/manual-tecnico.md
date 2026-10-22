# Manual tecnico

**Autor:** Christian Javier Rivas Arreaga, carné **202303204**. País asignado por último dígito (4): **CHN**.

**Estado (21 de octubre de 2026):** implementación y despliegue en GKE verificados. El flujo completo, Zot, Gateway API, operador, carga con Locust, Grafana y alerta de Falco se documentan en [evidencia-gcp.md](evidencia-gcp.md).

## 1. Descripcion general

M.U.M.N.K8s es un sistema distribuido para recibir, transportar, procesar y almacenar reportes militares simulados. El proyecto usa servicios independientes para representar un flujo de eventos en tiempo real dentro de Kubernetes.

Cada reporte tiene la siguiente estructura:

```json
{
  "country": "GTM",
  "warplanes_in_air": 42,
  "warships_in_water": 14,
  "timestamp": "2026-10-07T23:40:00Z"
}
```

Los rangos aplicados por la API son:

| Campo | Regla |
|---|---|
| `country` | `USA`, `RUS`, `CHN`, `ESP` o `GTM` |
| `warplanes_in_air` | Entero entre 0 y 50 |
| `warships_in_water` | Entero entre 0 y 30 |
| `timestamp` | Fecha valida en formato RFC3339 |

## 2. Arquitectura

```text
Cliente Locust
      |
      v
Kubernetes Gateway API
      |
      v
Service api-rust
      |
      v
API Rust :8080
      |
      | HTTP interno
      v
Service go-ingest :8081
      |
      | HTTP localhost:8082
      v
Cliente gRPC (segundo contenedor de go-ingest)
      |
      | protobuf/gRPC
      v
Service go-writer :50051
      |
      | HTTP localhost:8083
      v
Publicador RabbitMQ (segundo contenedor de go-writer)
      |
      | AMQP
      v
RabbitMQ
      |
      | Consumer acknowledgement
      v
Consumer Go
      |
      | RESP
      v
Valkey :6379
      |
      v
Exporter de metricas :9091 -> Prometheus :9090 -> Grafana
```

El despliegue activo usa GKE Standard zonal y Zot HTTPS en una VM N1 separada. Véase [despliegue-gcp.md](despliegue-gcp.md) para recursos y comandos reproducibles.

## 3. Componentes

### 3.1. API Rust

Ubicacion: `api-rust/`.

Responsabilidades:

- Exponer el endpoint publico.
- Validar la estructura del reporte.
- Reenviar el reporte al servicio Go.
- Responder con `202 Accepted` cuando el servicio interno acepta el reporte.
- Exponer `/health` para probes de Kubernetes.

Endpoint:

```text
POST /grpc-202303204
```

Aunque el nombre de la ruta identifica el flujo gRPC solicitado por el enunciado, la entrada publica es HTTP REST. El intercambio gRPC ocurre entre los servicios Go.

### 3.2. Go Ingest

Ubicacion: `go-ingest/`.

Responsabilidades:

- Recibir el reporte desde Rust.
- Validar que los campos basicos esten presentes.
- Reenviar el JSON al contenedor cliente gRPC del mismo pod.
- El contenedor `go-ingest-client` convierte el JSON al contrato protobuf e invoca `SendReport` en `go-writer`.

Endpoint interno:

```text
POST /reports
```

### 3.3. Go Writer

Ubicacion: `go-writer/`.

Responsabilidades:

- Implementar el servidor gRPC.
- Recibir `SendReport` en el contenedor servidor gRPC.
- Serializar la solicitud y enviarla al contenedor publicador del mismo pod.
- El contenedor `go-writer-publisher` declara la cola durable `military-reports`, publica mensajes persistentes y espera confirmacion de RabbitMQ.
- Devolver confirmacion al cliente gRPC tras recibirla del broker.

### 3.4. Consumer

Ubicacion: `consumer/`.

Responsabilidades:

- Consumir mensajes de `military-reports`.
- Deserializar cada reporte.
- Guardarlo en Valkey mediante RESP.
- Confirmar el mensaje solo despues de guardarlo correctamente.
- Rechazar o reencolar mensajes segun el tipo de error.

Las claves se generan como:

```text
reports:USA
reports:RUS
reports:CHN
reports:ESP
reports:GTM
```

Cada reporte se guarda como un elemento de una lista mediante `RPUSH`.

### 3.5. RabbitMQ

RabbitMQ es el broker obligatorio del flujo principal. Se prepara como StatefulSet con PVC de 2 GiB. La cola se declara durable y los mensajes se publican en modo persistente. El usuario y la contraseña se obtienen del Secret `rabbitmq-credentials`, creado fuera de Git antes de desplegar.

### 3.6. Valkey

Valkey se despliega como StatefulSet con PVC. La configuracion incluye:

- AOF para persistencia.
- Limite de memoria de 256 MB.
- Politica `allkeys-lru` para evitar saturacion.
- Service interno en el puerto 6379.

### 3.7. Grafana

Grafana se administra mediante un chart Helm. El chart contiene:

- Deployment.
- Service.
- Referencia al Secret externo `grafana-admin` para el acceso administrativo.
- Configuracion de datasource.
- Provisioning de dashboards.

El datasource integrado de Prometheus apunta a `prometheus:9090`. El exporter `metrics` lee las listas `reports:*` de Valkey y publica `/metrics` con máximos, mínimos, modas, totales por país y la última muestra de CHN. Prometheus conserva la serie de muestras para las gráficas temporales. El dashboard provisionado tiene doce paneles y usa CHN porque el carné termina en 4. En GKE se verificaron el dashboard provisionado y las series con datos reales; la inspección visual de los doce paneles durante la defensa sigue siendo recomendable.

## 4. Comunicacion gRPC

El contrato se encuentra en `proto/report.proto`.

El metodo principal es:

```text
SendReport(WarReportRequest) returns (WarReportResponse)
```

La solicitud incluye pais, aviones, barcos y timestamp. La respuesta informa si el mensaje fue publicado.

El servicio gRPC no se expone al exterior del cluster. Solo se publica el endpoint de Rust mediante Gateway API.

## 5. Gateway API

El recurso `Gateway` utiliza la clase:

```text
gke-l7-global-external-managed
```

El recurso `HTTPRoute` enruta el prefijo `/grpc-202303204` al Service `api-rust` en el puerto 80. Esta configuracion reemplaza el uso de un Ingress Controller, tal como solicita el enunciado.

## 6. Escalamiento y recursos

La API Rust tiene un HPA con:

| Parametro | Valor |
|---|---:|
| Minimo de replicas | 1 |
| Maximo de replicas | 3 |
| Recurso medido | CPU |
| Objetivo | 30% |

Todos los servicios principales tienen requests y limits. Esto permite que Kubernetes programe los pods correctamente y que el HPA pueda calcular el uso relativo de CPU.

En la carga principal de GKE, el HPA de Rust pasó de una a dos réplicas. También se compararon los deployments Go con una y dos réplicas usando 20 usuarios, rampa de 5 usuarios/s y 45 segundos por ejecución. La tasa fue 47,88 frente a 47,90 solicitudes/s, ambas sin fallos; el p95 fue 130 frente a 140 ms. Dos réplicas no mejoraron esa latencia en la carga corta. Los datos acumulados no se reiniciaron, así que el resultado no permite inferir una ventaja general. Detalles en [evidencia-gcp.md](evidencia-gcp.md).

## 7. Operator Helm-based

El Operator contempla dos recursos personalizados:

- `ValkeyInstance`.
- `GrafanaInstance`.

Los CRDs se encuentran en `k8s/07-crds.yaml` y los recursos de ejemplo en `k8s/06-custom-resources.yaml`.

El chart de Valkey crea un StatefulSet, un Service y un PVC. El chart de Grafana crea un Deployment, un Service y un ConfigMap con datasource y dashboard; lee la contraseña administrativa del Secret externo `grafana-admin`.

La carpeta `operator/` contiene `Dockerfile`, `watches.yaml` y ambos charts. `k8s/09-operator.yaml` aporta el ServiceAccount, RBAC y Deployment. En kind se comprobó instalación, actualización y eliminación de CRs; en GKE ambos CR se reconciliaron y los workloads quedaron listos. El scaffold oficial de Operator SDK v1.37.0 se generó con `init` y `create api` y sus archivos `PROJECT`, `Makefile` y `config/` están en `operator/`.

## 8. Seguridad con Falco

La regla personalizada detecta intentos de lectura de:

- `/etc/shadow`.
- `/app/secretos`.

La regla se limita a contenedores cuyas imagenes pertenecen al proyecto. El procedimiento de instalacion y la evidencia esperada estan en `docs/falco.md`.

Regla personalizada preparada en `k8s/falco-values.yaml`:

```yaml
- rule: MUMNK8s sensitive file read
  desc: Detecta lectura de archivos sensibles en pods del proyecto
  condition: >
    open_read and container and
    (fd.name=/etc/shadow or fd.name startswith /app/secretos) and
    (container.image.repository contains "mumnk8s")
  output: >
    MUMNK8s intento de leer archivo sensible
    (user=%user.name command=%proc.cmdline file=%fd.name container=%container.name)
  priority: WARNING
  tags: [mumnk8s, filesystem]
```

La alerta real se produjo en GKE al leer `/app/secretos/comprobacion` desde `rabbitmq-0`; el registro incluyó proceso, archivo, contenedor y pod. El procedimiento reproducible está en [falco.md](falco.md).

## 9. HPA, Locust y evidencias

Cada prueba debe registrar:

1. Configuracion de usuarios y tiempo de ejecucion.
2. Solicitudes por segundo.
3. Porcentaje de errores.
4. Latencia promedio y percentiles.
5. Numero de replicas antes y despues.
6. Uso de CPU.
7. Estado de RabbitMQ.
8. Reportes almacenados en Valkey.

Los resultados medidos y los comandos de comprobación están en [evidencia-gcp.md](evidencia-gcp.md). Durante la revisión se pueden capturar las pantallas actuales de GKE, Grafana y Falco; este repositorio conserva evidencia textual reproducible.

## 10. Limitaciones conocidas

- Zot requiere un dominio HTTPS real y confianza del certificado en los nodos.
- Gateway API depende de la disponibilidad de la GatewayClass en GKE.
- HPA requiere metricas disponibles mediante Metrics Server.
- El dashboard está provisionado en GKE y recibe métricas reales; se debe mantener el túnel `port-forward` abierto para verlo desde Windows.
- El contrato `.proto` genera `report.pb.go` y `report_grpc.pb.go`; cliente y servidor usan protobuf estándar. El writer convierte la solicitud a JSON con nombres de campo del proto antes de publicarla en RabbitMQ.
- Los deployments `go-ingest` y `go-writer` incluyen dos contenedores cada uno; el flujo de GKE respondió sin fallos durante la carga medida.
- La ruta `/dapr-202303204` es opcional y no forma parte del flujo base.
- La infraestructura consume crédito mientras siga desplegada; el presupuesto configurado envía alertas pero no detiene recursos.

## 11. Pruebas realizadas y conclusiones provisionales

El 19 de octubre de 2026 se ejecutaron las pruebas Go de todos los módulos y `go vet`; ambas validaciones finalizaron sin errores. Las cuatro pruebas unitarias de la API Rust pasaron. Ambos charts pasaron `helm lint` y `helm template`; los 29 recursos de Kubernetes se leyeron sin identidades duplicadas. El dashboard JSON contiene doce paneles.

También se construyeron las ocho imágenes con Docker Engine en WSL Ubuntu. Docker Compose confirmó el flujo API Rust → Go → RabbitMQ → consumer → Valkey; un reporte CHN quedó almacenado, Prometheus leyó sus métricas y Grafana cargó el dashboard provisionado. La primera prueba de carga reveló un problema de arranque del publicador; tras añadir reinicio automático en Compose, Locust ejecutó 1857 solicitudes en un minuto con 10 usuarios y 0 fallos (mediana de 8 ms, percentil 95 de 14 ms). Estos números describen solo el equipo local.

En kind se instaló el Helm Operator. Creó Valkey y Grafana desde CRs separados; actualizar `spec.maxMemory` cambió el StatefulSet de Valkey y el dato CHN sobrevivió al reinicio. También se creó y eliminó una instancia Valkey de prueba. Envoy Gateway recibió reportes mediante la HTTPRoute. Metrics Server alimentó al HPA, que aumentó la API Rust de una a tres réplicas durante 90 segundos de Locust: 10873 solicitudes, 0 fallos, mediana de 480 ms y percentil 95 de 660 ms. La cola RabbitMQ se drenó tras la prueba. El Gateway local no obtuvo IP externa porque kind no proporciona un LoadBalancer; se comprobó por `port-forward`.

El 21 de octubre se verificó Zot HTTPS en una VM N1, la descarga de imágenes y un artefacto OCI, GKE con un nodo N1 listo, Gateway API `Programmed=True` y la ruta pública con respuesta `202`. Locust realizó 6569 POST en 90 segundos con 50 usuarios, cero fallos, 74,58 solicitudes/s, mediana 240 ms y p95 810 ms. El HPA de Rust pasó de una a dos réplicas y la cola RabbitMQ quedó vacía. Prometheus devolvió 6574 reportes totales y 1314 de CHN; Grafana presentó el dashboard de doce paneles. Falco generó la alerta personalizada. Una comparación adicional de una y dos réplicas de los servicios Go arrojó rendimiento similar con 20 usuarios. Son resultados puntuales de GKE y no sustituyen la evaluación presencial. Los comandos y condiciones se detallan en [evidencia-gcp.md](evidencia-gcp.md).
