# Artifact OCI en Zot

## Objetivo

El enunciado solicita descargar un archivo de entrada desde un registry como artifact OCI. En este proyecto se utiliza `k8s/00-namespace.yaml` para demostrar el flujo de publicacion y descarga.

## Publicar el artifact

Definir el registry real:

```bash
export ZOT=registry.example.com
```

Publicar el archivo con ORAS:

```bash
oras push "$ZOT/mumnk8s/input:dev" \
  k8s/00-namespace.yaml:application/yaml
```

Verificar el artifact:

```bash
oras manifest fetch "$ZOT/mumnk8s/input:dev"
```

## Descargar y utilizar el artifact

```bash
mkdir -p /tmp/mumnk8s-input
oras pull "$ZOT/mumnk8s/input:dev" -o /tmp/mumnk8s-input
kubectl apply -f /tmp/mumnk8s-input/00-namespace.yaml
```

El artifact es un archivo de configuracion. No sustituye las imagenes de los servicios ni contiene credenciales. El registry debe utilizar HTTPS y sus permisos deben configurarse antes de publicar.
