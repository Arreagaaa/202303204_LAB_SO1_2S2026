# M.U.M.N.K8s

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
Go Ingest
  |
  | gRPC
  v
Go Writer
  |
  v
RabbitMQ
  |
  v
Consumer Go
  |
  v
Valkey
  |
  v
Grafana
```

## Estado del proyecto

La implementacion actual cubre la base funcional y de despliegue del flujo principal. Los componentes que dependen de una cuenta de GCP, una VM con Zot o un cluster GKE deben validarse en el ambiente correspondiente.

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
| CRDs y Custom Resources | Preparados |
| Regla personalizada de Falco | Preparada |
| Publicacion de imagenes en Zot | Requiere ambiente GCP |
| Despliegue final en GKE | Requiere ambiente GCP |
| Ruta opcional con Dapr | Pendiente |

## Estructura del repositorio

| Ruta | Responsabilidad |
|---|---|
| `api-rust/` | API publica y validacion de reportes |
| `go-ingest/` | Endpoint interno y cliente gRPC |
| `go-writer/` | Servidor gRPC y publicador RabbitMQ |
| `consumer/` | Consumidor RabbitMQ y persistencia en Valkey |
| `proto/` | Contrato del servicio gRPC |
| `k8s/` | Namespace, deployments, services, Gateway, HPA, CRDs y reglas |
| `operator/` | Charts Helm para Valkey y Grafana |
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
