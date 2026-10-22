# Instalación y prueba de Falco

## Objetivo

Falco observa llamadas al sistema dentro de los contenedores. Aquí detecta la lectura de archivos sensibles desde los pods de la aplicación.

## Regla del proyecto

La regla `MUMNK8s sensitive file read` genera una alerta cuando un contenedor del proyecto intenta leer:

- `/etc/shadow`.
- `/proc/1/environ` (variables de entorno del proceso principal).
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

Ejecutar una lectura desde el pod de la API Rust, sin imprimir el contenido sensible:

```bash
kubectl exec -n mumnk8s deploy/api-rust -- sh -c 'cat /proc/1/environ >/dev/null'
```

Consultar los logs:

```bash
kubectl logs -n falco daemonset/falco --since=2m | grep 'MUMNK8s intento de leer archivo sensible'
```

El 22 de octubre de 2026, tras actualizar Falco a la revisión Helm 2, GKE registró `command=cat /proc/1/environ`, `file=/proc/1/environ`, `container=api-rust`, `k8s_pod_name=api-rust-5fd9b8d4b8-5ljtb` y `k8s_ns_name=mumnk8s`. La regla y el DaemonSet quedaron listos. La prueba previa desde RabbitMQ permanece como antecedente, pero esta ejecución cumple la condición de usar un pod Rust o Go. Falta únicamente la captura que incorporará el estudiante al manual antes de entregar.
