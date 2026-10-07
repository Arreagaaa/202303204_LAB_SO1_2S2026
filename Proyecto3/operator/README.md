# Operator M.U.M.N.K8s

El Operator se construye con Operator SDK y el plugin Helm. Cada recurso personalizado representa una instancia administrada:

- `ValkeyInstance` instala Valkey con almacenamiento persistente.
- `GrafanaInstance` instala Grafana con el datasource y los dashboards del proyecto.

Los charts que se deben encapsular en el proyecto generado por Operator SDK se encuentran en `helm-charts/`. El reconciliador Helm mantiene los recursos creados por cada Custom Resource y permite cambiar replicas, almacenamiento y puerto desde YAML.

## Inicializar el proyecto

```bash
operator-sdk init --plugins helm
operator-sdk create api --group mumnk8s --version v1alpha1 --kind ValkeyInstance
operator-sdk create api --group mumnk8s --version v1alpha1 --kind GrafanaInstance
```

Luego se copian los charts de esta carpeta dentro de `helm-charts/` del operador generado y se actualizan los valores desde los campos del recurso personalizado.

## Instalacion

```bash
make docker-build docker-push IMG=$ZOT/mumnk8s/operator:dev
make deploy IMG=$ZOT/munmk8s/operator:dev
kubectl apply -f ../k8s/06-custom-resources.yaml
```
