# Manual tecnico

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
      | gRPC
      v
Service go-writer :50051
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
Grafana
```

El cluster se ejecuta en GKE. Zot se aloja en una VM N1 separada y funciona como registry privado para las imagenes del proyecto.

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
POST /grpc-2023204
```

Aunque el nombre de la ruta identifica el flujo gRPC solicitado por el enunciado, la entrada publica es HTTP REST. El intercambio gRPC ocurre entre los servicios Go.

### 3.2. Go Ingest

Ubicacion: `go-ingest/`.

Responsabilidades:

- Recibir el reporte desde Rust.
- Validar que los campos basicos esten presentes.
- Convertir el JSON a la estructura del contrato gRPC.
- Invocar `SendReport` en `go-writer`.

Endpoint interno:

```text
POST /reports
```

### 3.3. Go Writer

Ubicacion: `go-writer/`.

Responsabilidades:

- Implementar el servidor gRPC.
- Declarar la cola durable `military-reports`.
- Serializar cada solicitud recibida.
- Publicarla en RabbitMQ.
- Devolver confirmacion al cliente gRPC.

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

RabbitMQ es el broker obligatorio del flujo principal. La cola se declara durable para conservar la configuracion ante reinicios del broker.

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
- Secret de acceso administrativo.
- Configuracion de datasource.
- Provisioning de dashboards.

El datasource utiliza el plugin Redis y apunta al Service de Valkey.

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

El recurso `HTTPRoute` enruta el prefijo `/grpc-2023204` al Service `api-rust` en el puerto 80. Esta configuracion reemplaza el uso de un Ingress Controller, tal como solicita el enunciado.

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

El chart de Valkey crea un StatefulSet, un Service y un PVC. El chart de Grafana crea un Deployment, un Service, un Secret y un ConfigMap con datasource y dashboard.

La carpeta `operator/` contiene los charts base que se deben integrar al proyecto generado por Operator SDK. La reconciliacion final debe comprobarse dentro del cluster observando los recursos creados despues de aplicar cada Custom Resource.

## 8. Seguridad con Falco

La regla personalizada detecta intentos de lectura de:

- `/etc/shadow`.
- `/app/secretos`.

La regla se limita a contenedores cuyas imagenes pertenecen al proyecto. El procedimiento de instalacion y la evidencia esperada estan en `docs/falco.md`.

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
- Los dashboards de Grafana necesitan consultas finales contra el datasource instalado.
- La ruta `/dapr-2023204` es opcional y no forma parte del flujo base.
- Las capturas de GCP, Zot, GKE, Locust, HPA y Falco deben generarse en el ambiente real de pruebas.
