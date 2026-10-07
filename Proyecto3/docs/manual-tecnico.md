# Manual tecnico inicial

## Arquitectura

La API Rust recibe reportes JSON por `/grpc-2023204`. Valida los rangos definidos por el enunciado y envia el reporte al servicio `go-ingest`. Este servicio expone un endpoint interno y funciona como cliente gRPC. El servidor gRPC de `go-writer` publica cada mensaje en la cola durable `military-reports` de RabbitMQ.

El consumidor confirma cada mensaje solamente despues de guardarlo en Valkey. La clave usa el pais del reporte y cada valor conserva el JSON recibido. La politica de memoria de Valkey evita que un crecimiento accidental deje sin memoria al pod.

## Comunicacion

El Gateway API solamente expone el prefijo `/grpc-2023204`. El resto de servicios se comunica por Services internos del namespace `mumnk8s`. El cliente y servidor gRPC usan el contrato de `proto/report.proto`; el transporte queda dentro del cluster.

## Escalamiento

La API Rust tiene requests y limits de CPU. El HPA permite de una a tres replicas y usa un objetivo del 30 por ciento, como solicita el enunciado. Las pruebas de Locust deben repetirse con una y dos replicas de los servicios Go para comparar resultados.

## Operator

El Operator Helm-based administra los recursos `ValkeyInstance` y `GrafanaInstance`. Cada recurso representa una instalacion declarativa y sus valores se transforman en valores del chart correspondiente. Esta entrega deja los charts y los Custom Resources listos para integrarse al esqueleto generado por Operator SDK.

## Evidencias pendientes

Para la entrega en GCP se deben agregar capturas del cluster GKE, la VM de Zot, la publicacion de imagenes, el estado del Gateway, las pruebas de Locust, los cambios del HPA y los dashboards de Grafana. Tambien se debe documentar el archivo OCI descargado desde Zot cuando se defina el archivo de entrada.
