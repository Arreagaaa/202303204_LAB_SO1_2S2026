# Guia de instalacion

## Requisitos

Se necesita Docker, Go 1.22, Rust, kubectl, Helm, Locust y acceso a un cluster GKE. Las imagenes finales se publican en Zot usando HTTPS.

## Pruebas locales

Desde `Proyecto3` se descargan las dependencias y se compilan los servicios:

```bash
go work sync
cargo build --manifest-path api-rust/Cargo.toml
```

El contrato `proto/report.proto` documenta el mensaje gRPC. El archivo `proto/proto.go` contiene el cliente y el servidor usados por esta primera version.

RabbitMQ y Valkey pueden levantarse localmente:

```bash
docker run -d --name rabbitmq -p 5672:5672 rabbitmq:3.13
docker run -d --name valkey -p 6379:6379 valkey/valkey:8
```

## Construccion de imagenes

Los Dockerfiles esperan que el contexto sea la carpeta `Proyecto3`:

```bash
docker build -f api-rust/Dockerfile -t $ZOT/mumnk8s/api-rust:dev .
docker build -f go-ingest/Dockerfile -t $ZOT/mumnk8s/go-ingest:dev .
docker build -f go-writer/Dockerfile -t $ZOT/mumnk8s/go-writer:dev .
docker build -f consumer/Dockerfile -t $ZOT/mumnk8s/consumer:dev .
docker push $ZOT/mumnk8s/api-rust:dev
docker push $ZOT/mumnk8s/go-ingest:dev
docker push $ZOT/mumnk8s/go-writer:dev
docker push $ZOT/mumnk8s/consumer:dev
```

`ZOT` debe ser el dominio del registry con su certificado HTTPS configurado.

## Despliegue

```bash
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/01-rabbitmq.yaml
kubectl apply -f k8s/02-services.yaml
kubectl apply -f k8s/03-hpa.yaml
kubectl apply -f k8s/04-gateway.yaml
kubectl apply -f k8s/05-consumer-valkey.yaml
kubectl apply -f k8s/07-crds.yaml
```

El Operator se instala con el proyecto generado por Operator SDK y despues se aplican los recursos personalizados:

```bash
kubectl apply -f k8s/06-custom-resources.yaml
```

## Prueba de carga

Cuando el Gateway tenga IP publica:

```bash
locust -f locust/locustfile.py --host http://IP_DEL_GATEWAY
```
