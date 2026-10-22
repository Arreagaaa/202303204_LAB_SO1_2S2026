# Operator Helm para Valkey y Grafana

El contenedor usa la imagen `helm-operator` de Operator SDK, `watches.yaml` y dos charts. Cada Custom Resource activa una reconciliacion independiente. El proyecto incluye Dockerfile, watches, charts, CRDs, RBAC y Deployment del operador. En kind se comprobaron creacion, actualizacion y eliminacion de una instancia Valkey, ademas de la instalacion de Grafana. En GKE ambos CR se reconciliaron y sus workloads estan listos.

El 21 de octubre de 2026 se ejecuto Operator SDK v1.37.0 en un directorio temporal usando los charts existentes:

```bash
operator-sdk init --plugins helm --domain sopes1.usac --project-name mumnk8s-operator \
  --group mumnk8s --version v1alpha1 --kind ValkeyInstance \
  --helm-chart /ruta/absoluta/operator/helm-charts/valkeyinstance
operator-sdk create api --group mumnk8s --version v1alpha1 \
  --kind GrafanaInstance \
  --helm-chart /ruta/absoluta/operator/helm-charts/grafanainstance
```

Se incorporaron `PROJECT`, `Makefile` y `config/` generados. Los charts, `watches.yaml` y Dockerfile funcionales se conservaron. El RBAC generado por SDK se revisa como referencia; el despliegue activo usa `../k8s/09-operator.yaml`.

| Kind | Chart | Objetos derivados |
|---|---|---|
| `ValkeyInstance` | `helm-charts/valkeyinstance` | StatefulSet, Service y PVC |
| `GrafanaInstance` | `helm-charts/grafanainstance` | Deployment, Service y ConfigMap; usa el Secret externo `grafana-admin` |

Los CRDs estan en `../k8s/07-crds.yaml` y las instancias de ejemplo en `../k8s/06-custom-resources.yaml`. El consumer apunta a `valkey-reports:6379`, nombre generado por el chart para el CR `reports`.

## Construir y desplegar

Desde la raiz `Proyecto3`, con `ZOT` configurado y autenticacion HTTPS preparada:

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

Para probar la reconciliacion, cambiar `spec.maxMemory` del ValkeyInstance y observar la actualizacion del StatefulSet. Cambiar `storage` requiere un procedimiento de expansion del PVC; no es una actualizacion ordinaria del StatefulSet. Para comprobar la eliminacion, usar una instancia de prueba aparte y borrar ese CR. La instancia Grafana debe tener un nombre distinto al de Valkey porque Helm usa el nombre del CR como nombre del release; en los manifiestos es `dashboard`.
