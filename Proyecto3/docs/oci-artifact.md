# Artefacto OCI en Zot

## Objetivo

El enunciado solicita descargar un archivo de entrada desde un registro como artefacto OCI. En este proyecto se utiliza `k8s/00-namespace.yaml` para demostrar la publicación y descarga.

## Publicar el artefacto

Definir el registry real:

```bash
export ZOT=registry.example.com
```

Publicar el archivo con ORAS:

```bash
oras push "$ZOT/mumnk8s/input:dev" \
  k8s/00-namespace.yaml:application/yaml
```

Verificar el artefacto:

```bash
oras manifest fetch "$ZOT/mumnk8s/input:dev"
```

## Descargar y utilizar el artefacto

```bash
mkdir -p /tmp/mumnk8s-input
oras pull "$ZOT/mumnk8s/input:dev" -o /tmp/mumnk8s-input
kubectl apply -f /tmp/mumnk8s-input/k8s/00-namespace.yaml
```

El artefacto es un archivo de configuración. No sustituye las imágenes de los servicios ni contiene credenciales. El registro debe usar HTTPS y exigir autenticación antes de publicar.
