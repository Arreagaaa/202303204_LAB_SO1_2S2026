# Operator Helm para Valkey y Grafana

El operador usa la imagen `helm-operator` de Operator SDK, `watches.yaml` y dos charts. Cada recurso personalizado se reconcilia por separado. El proyecto incluye Dockerfile, watches, charts, CRDs, RBAC y Deployment. En kind se comprobaron la creación, actualización y eliminación de una instancia Valkey, además de la instalación de Grafana. En GKE ambos recursos se reconciliaron y sus componentes quedaron listos.

El 21 de octubre de 2026 se ejecutó Operator SDK v1.37.0 en un directorio temporal con los charts existentes:

```bash
operator-sdk init --plugins helm --domain sopes1.usac --project-name mumnk8s-operator \
  --group mumnk8s --version v1alpha1 --kind ValkeyInstance \
  --helm-chart /ruta/absoluta/operator/helm-charts/valkeyinstance
operator-sdk create api --group mumnk8s --version v1alpha1 \
  --kind GrafanaInstance \
  --helm-chart /ruta/absoluta/operator/helm-charts/grafanainstance
```

Se incorporaron los archivos `PROJECT`, `Makefile` y `config/` generados. Los charts, `watches.yaml` y el Dockerfile se conservaron. El RBAC generado por el SDK sirve como referencia; el despliegue activo usa `../k8s/09-operator.yaml`.

| Kind | Chart | Objetos derivados |
|---|---|---|
| `ValkeyInstance` | `helm-charts/valkeyinstance` | StatefulSet, Service y PVC |
| `GrafanaInstance` | `helm-charts/grafanainstance` | Deployment, Service y ConfigMap; usa el Secret externo `grafana-admin` |

Los CRDs están en `../k8s/07-crds.yaml` y las instancias de ejemplo en `../k8s/06-custom-resources.yaml`. El consumidor usa `valkey-reports:6379`, nombre que genera el chart para el recurso `reports`.

## Construir y desplegar

Desde la raíz `Proyecto3`, con `ZOT` configurado y la autenticación HTTPS preparada:

```bash
docker build -f operator/Dockerfile -t "$ZOT/mumnk8s/operator:dev" operator
docker push "$ZOT/mumnk8s/operator:dev"
sed -i "s#zot.example#$ZOT#g" k8s/09-operator.yaml
kubectl apply -f k8s/07-crds.yaml
kubectl apply -f k8s/09-operator.yaml
kubectl rollout status deployment/mumnk8s-operator -n mumnk8s
kubectl apply -f k8s/06-custom-resources.yaml
```

```bash
kubectl get valkeyinstances,grafanainstances -n mumnk8s
kubectl get statefulsets,deployments,services,pvc -n mumnk8s
kubectl logs deployment/mumnk8s-operator -n mumnk8s
kubectl exec -n mumnk8s statefulset/valkey-reports -- valkey-cli ping
```

Para probar la reconciliación, se cambia `spec.maxMemory` del ValkeyInstance y se observa la actualización del StatefulSet. Cambiar `storage` requiere expandir el PVC; no basta con actualizar el StatefulSet. Para comprobar la eliminación se usa una instancia de prueba y se borra ese recurso. La instancia de Grafana debe tener un nombre distinto al de Valkey porque Helm usa el nombre del recurso como nombre de la instalación; en los manifiestos se llama `dashboard`.
