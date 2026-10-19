# Operator Helm para Valkey y Grafana

El contenedor usa la imagen `helm-operator` de Operator SDK, `watches.yaml` y dos charts. Cada Custom Resource activa una reconciliacion independiente. El proyecto incluye Dockerfile, watches, charts, CRDs, RBAC y Deployment del operador. En kind se comprobaron creacion, actualizacion y eliminacion de una instancia Valkey, ademas de la instalacion de Grafana. Falta repetir la prueba en GKE.

El esqueleto se armó con el runtime oficial y la estructura del ejemplo de Semana 11, pero no se generó con el comando `operator-sdk init/create api` en este equipo. Si se exige evidencia literal de esos comandos, habrá que regenerar el esqueleto y conservar los charts y CRDs de este repositorio.

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
