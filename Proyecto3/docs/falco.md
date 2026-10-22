# Instalación y prueba de Falco

## Objetivo

Falco observa llamadas al sistema dentro de los contenedores. Aquí detecta la lectura de archivos sensibles desde los pods de la aplicación.

## Regla del proyecto

La regla `MUMNK8s sensitive file read` genera una alerta cuando un contenedor del proyecto intenta leer:

- `/etc/shadow`.
- Cualquier ruta que empiece con `/app/secretos`.

La definición versionada está en `k8s/falco-values.yaml`. El archivo `k8s/08-falco-rule.yaml` conserva la misma regla como ConfigMap de referencia.

## Instalación

```bash
bash scripts/gcp/deploy-falco.sh "$ZOT_HOST"
```

Comprobar el despliegue:

```bash
kubectl get pods -n falco
kubectl get daemonset -n falco
```

## Simulación

Ejecutar una lectura desde un pod de la aplicación:

```bash
kubectl exec -n mumnk8s rabbitmq-0 -- sh -c \
  'mkdir -p /app/secretos && echo prueba > /app/secretos/comprobacion && cat /app/secretos/comprobacion'
```

Consultar los logs:

```bash
kubectl logs -n falco daemonset/falco --since=2m | grep 'MUMNK8s intento de leer archivo sensible'
kubectl exec -n mumnk8s rabbitmq-0 -- rm -f /app/secretos/comprobacion
```

La prueba de GKE registró `command=cat`, `file=/app/secretos/comprobacion`, `container=rabbitmq`, `k8s_pod_name=rabbitmq-0` y `k8s_ns_name=mumnk8s`. La lectura de `/etc/shadow` desde la API Rust fue denegada por permisos, por lo que esa prueba no produjo una alerta.
