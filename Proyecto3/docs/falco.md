# Instalacion y prueba de Falco

## Objetivo

Falco permite observar llamadas al sistema realizadas desde los contenedores. En este proyecto se utiliza para detectar la lectura de archivos sensibles desde los pods de la aplicacion.

## Regla del proyecto

La regla `MUMNK8s sensitive file read` genera una alerta cuando un contenedor del proyecto intenta leer:

- `/etc/shadow`.
- Cualquier ruta que empiece con `/app/secretos`.

La definicion versionada esta en `k8s/falco-values.yaml`. El archivo `k8s/08-falco-rule.yaml` conserva la misma regla como ConfigMap de referencia.

## Instalacion

```bash
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update
helm upgrade --install falco falcosecurity/falco \
  --namespace falco \
  --create-namespace \
  --values k8s/falco-values.yaml
```

Comprobar el despliegue:

```bash
kubectl get pods -n falco
kubectl get daemonset -n falco
```

## Simulacion

Ejecutar una lectura desde un pod de la aplicacion:

```bash
kubectl exec -n mumnk8s deploy/api-rust -- cat /etc/shadow
```

Consultar los logs:

```bash
kubectl logs -n falco daemonset/falco --since=2m | grep MUMNK8s
```

La evidencia debe mostrar el nombre de la regla, el proceso, el archivo y el contenedor involucrado.
