# Proyecto 3: M.U.M.N.K8s

Este directorio contiene la primera parte del proyecto de monitoreo de unidades militares. El flujo principal es:

Locust -> Gateway API -> API Rust -> servicio Go -> gRPC -> RabbitMQ -> consumidor Go -> Valkey -> Grafana

La implementacion se separa en servicios pequenos para que cada componente pueda probarse y desplegarse por separado.

## Estado del avance

Esta primera entrega incluye:

- API REST en Rust con el endpoint `/grpc-2023204`.
- Servicio Go de entrada con endpoint REST y cliente gRPC.
- Servicio Go escritor con servidor gRPC y publicacion en RabbitMQ.
- Consumidor Go con almacenamiento en Valkey.
- Contrato protobuf compartido.
- Manifiestos base de Kubernetes, Gateway API, HPA, RabbitMQ y Locust.
- Base de Operator SDK basada en Helm para Valkey y Grafana.
- Guia de instalacion y manual tecnico inicial.

La ruta `/dapr-2023204` queda reservada para la implementacion opcional.

## Estructura

```text
api-rust/       API publica
go-ingest/      REST interno y cliente gRPC
go-writer/      servidor gRPC y publicador RabbitMQ
consumer/       consumidor RabbitMQ y cliente Valkey
proto/          contrato protobuf
k8s/            recursos del cluster
operator/       chart y recursos personalizados
locust/         prueba de carga
docs/           manual y guia de instalacion
```
