# Manual tecnico

**Autor:** Christian Javier Rivas Arreaga, carné **202303204**. País asignado por último dígito (4): **CHN**.

**Estado:** implementación de código y manifiestos preparada; compilación Go, validación Helm y pruebas de extremo a extremo en Compose y kind ejecutadas. Despliegue GKE y evidencias de carga en la nube pendientes.

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

El destino previsto es GKE. Zot deberá alojarse en una VM N1 separada y servir como registry HTTPS; ese despliegue no se ha ejecutado todavía.

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

El datasource integrado de Prometheus apunta a `prometheus:9090`. El exporter `metrics` lee las listas `reports:*` de Valkey y publica `/metrics` con máximos, mínimos, modas, totales por país y la última muestra de CHN. Prometheus conserva la serie de muestras para las gráficas temporales. El dashboard provisionado tiene doce paneles y usa CHN porque el carné termina en 4. Grafana local cargó el dashboard; falta revisar visualmente cada panel en GKE.

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

La comparacion de rendimiento debe ejecutarse con una y dos replicas de los deployments Go, manteniendo constantes los usuarios de Locust y la duracion de la prueba.

## 7. Operator Helm-based

El Operator contempla dos recursos personalizados:

- `ValkeyInstance`.
- `GrafanaInstance`.

Los CRDs se encuentran en `k8s/07-crds.yaml` y los recursos de ejemplo en `k8s/06-custom-resources.yaml`.

El chart de Valkey crea un StatefulSet, un Service y un PVC. El chart de Grafana crea un Deployment, un Service y un ConfigMap con datasource y dashboard; lee la contraseña administrativa del Secret externo `grafana-admin`.

La carpeta `operator/` contiene `Dockerfile`, `watches.yaml` y ambos charts. `k8s/09-operator.yaml` aporta el ServiceAccount, RBAC y Deployment. En kind se comprobó que el Helm Operator instala, actualiza y desinstala los recursos derivados del CR. La prueba debe repetirse en GKE.

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

La captura de la alerta real se añadira despues de ejecutar la simulacion en GKE.

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

Las capturas deben incorporarse al manual final junto con una explicacion breve de los resultados.

## 10. Limitaciones conocidas

- Zot requiere un dominio HTTPS real y confianza del certificado en los nodos.
- Gateway API depende de la disponibilidad de la GatewayClass en GKE.
- HPA requiere metricas disponibles mediante Metrics Server.
- El dashboard está provisionado en Grafana local y las métricas muestran datos reales; verificar visualmente todos los paneles y su comportamiento en GKE.
- El contrato `.proto` genera `report.pb.go` y `report_grpc.pb.go`; cliente y servidor usan protobuf estándar. El writer convierte la solicitud a JSON con nombres de campo del proto antes de publicarla en RabbitMQ.
- Los deployments `go-ingest` y `go-writer` incluyen dos contenedores cada uno. Falta validar su comunicación bajo carga en GKE.
- La ruta `/dapr-202303204` es opcional y no forma parte del flujo base.
- Las capturas de GCP, Zot, GKE, Locust, HPA y Falco deben generarse en el ambiente real de pruebas.

## 11. Pruebas realizadas y conclusiones provisionales

El 19 de octubre de 2026 se ejecutaron las pruebas Go de todos los módulos y `go vet`; ambas validaciones finalizaron sin errores. Las cuatro pruebas unitarias de la API Rust pasaron. Ambos charts pasaron `helm lint` y `helm template`; los 29 recursos de Kubernetes se leyeron sin identidades duplicadas. El dashboard JSON contiene doce paneles.

También se construyeron las ocho imágenes con Docker Engine en WSL Ubuntu. Docker Compose confirmó el flujo API Rust → Go → RabbitMQ → consumer → Valkey; un reporte CHN quedó almacenado, Prometheus leyó sus métricas y Grafana cargó el dashboard provisionado. La primera prueba de carga reveló un problema de arranque del publicador; tras añadir reinicio automático en Compose, Locust ejecutó 1857 solicitudes en un minuto con 10 usuarios y 0 fallos (mediana de 8 ms, percentil 95 de 14 ms). Estos números describen solo el equipo local.

En kind se instaló el Helm Operator. Creó Valkey y Grafana desde CRs separados; actualizar `spec.maxMemory` cambió el StatefulSet de Valkey y el dato CHN sobrevivió al reinicio. También se creó y eliminó una instancia Valkey de prueba. Envoy Gateway recibió reportes mediante la HTTPRoute. Metrics Server alimentó al HPA, que aumentó la API Rust de una a tres réplicas durante 90 segundos de Locust: 10873 solicitudes, 0 fallos, mediana de 480 ms y percentil 95 de 660 ms. La cola RabbitMQ se drenó tras la prueba. El Gateway local no obtuvo IP externa porque kind no proporciona un LoadBalancer; se comprobó por `port-forward`.

Faltan la prueba de Zot, la GatewayClass de GKE y la comparación controlada de una y dos réplicas en GKE. Estas mediciones de kind describen solo el equipo local.
