# M.U.M.N.K8s

**Autor:** Christian Javier Rivas Arreaga · **Carné:** 202303204 · **País asignado:** CHN.

## Monitoreo de Unidades Militares en la Nube con Kubernetes

Proyecto individual del curso Sistemas Operativos 1. La solucion simula la recepcion y el procesamiento de reportes militares generados desde distintos paises. Cada reporte contiene el pais de origen, la cantidad de aviones en el aire, la cantidad de barcos en el agua y la fecha del evento.

La arquitectura utiliza una API REST en Rust, servicios internos en Go, comunicacion gRPC, RabbitMQ, Valkey, Grafana, Kubernetes Gateway API, HPA y Locust. El despliegue objetivo es Google Kubernetes Engine (GKE) y las imagenes se publican en un registry Zot externo al cluster.

## Flujo principal

```text
Locust
  |
  v
Gateway API
  |
  v
API REST en Rust
  |
  v
Go Ingest REST + cliente gRPC (dos contenedores)
  |
  | gRPC
  v
Go Server gRPC + publicador RabbitMQ (dos contenedores)
  |
  v
RabbitMQ
  |
  v
Consumer Go
  |
  v
Valkey -> exporter -> Prometheus
  |
  v
Grafana
```

## Estado del proyecto

El flujo completo funciona en GKE: Zot sirve las imagenes, Gateway enruta a la API, el operador reconcilia Valkey y Grafana, y Prometheus alimenta los doce paneles. La carga de Locust y una alerta real de Falco estan documentadas en [docs/evidencia-gcp.md](docs/evidencia-gcp.md). La entrega en UEDI y la defensa del proyecto siguen siendo pasos del estudiante.

| Componente | Estado |
|---|---|
| API REST en Rust | Implementada |
| Validacion de reportes | Implementada |
| Cliente y servidor gRPC | Implementados |
| Publicacion en RabbitMQ | Implementada |
| Consumer y almacenamiento en Valkey | Implementados |
| Manifiestos de Kubernetes | Preparados |
| Gateway API | Preparado |
| HPA de Rust | Preparado |
| Prueba de carga con Locust | Preparada |
| Charts Helm de Valkey y Grafana | Preparados |
| Operator Helm, RBAC y watches | Funcionando en GKE; scaffold generado con Operator SDK |
| Exporter, Prometheus y dashboard de 12 paneles | Funcionando con datos reales en GKE |
| CRDs y Custom Resources | Preparados |
| Regla personalizada de Falco | Alerta comprobada en GKE |
| Publicacion de imagenes en Zot | Comprobada en HTTPS, incluido artefacto OCI |
| Despliegue final en GKE | Activo y validado |
| Ruta opcional con Dapr | Pendiente |

## Estructura del repositorio

| Ruta | Responsabilidad |
|---|---|
| `api-rust/` | API publica y validacion de reportes |
| `go-ingest/`, `go-ingest-client/` | Endpoint REST y cliente gRPC |
| `go-writer/`, `go-writer-publisher/` | Servidor gRPC y publicador RabbitMQ |
| `consumer/` | Consumidor RabbitMQ y persistencia en Valkey |
| `proto/` | Contrato del servicio gRPC |
| `k8s/` | Namespace, deployments, services, Gateway, HPA, CRDs y reglas |
| `operator/` | Helm Operator y charts para Valkey y Grafana |
| `metrics/` | Exporter Valkey para Prometheus |
| `locust/` | Generador de carga |
| `scripts/` | Validaciones reproducibles desde WSL2 |
| `docs/` | Documentacion de instalacion, arquitectura y evidencias |

## Inicio rapido

La guia completa se encuentra en [docs/guia-instalacion.md](docs/guia-instalacion.md). Para una validacion local desde WSL2:

```bash
cd Proyecto3
bash scripts/test-wsl.sh
```

Para comprender la arquitectura y el flujo de datos, consultar [docs/manual-tecnico.md](docs/manual-tecnico.md).

## Documentacion

- [Guia de instalacion](docs/guia-instalacion.md)
- [Manual tecnico](docs/manual-tecnico.md)
- [Operator Helm-based](operator/README.md)
- [Falco](docs/falco.md)
- [OCI Artifact en Zot](docs/oci-artifact.md)
- [Despliegue en GCP](docs/despliegue-gcp.md)
- [Evidencia de GCP](docs/evidencia-gcp.md)
