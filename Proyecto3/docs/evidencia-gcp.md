# Evidencia de despliegue en GCP

Verificado el 21 de octubre de 2026 (hora de Guatemala) en el proyecto `ID_DEL_PROYECTO`. Estos resultados son una medición puntual; se pueden repetir con [despliegue-gcp.md](despliegue-gcp.md).

## Infraestructura y flujo

| Elemento | Resultado observado |
|---|---|
| Zot | VM `mumnk8s-zot` N1 standard 1 en `us-central1-a`; HTTPS en `203-0-113-10.sslip.io` |
| Imágenes y artefactos | Imágenes de aplicación, operador, infraestructura y Falco publicadas en Zot; archivo `k8s/00-namespace.yaml` publicado y descargado como artefacto OCI |
| GKE | Clúster zonal Standard `mumnk8s-gke-c` en `us-central1-c`, un nodo N1 standard 2 `Ready` |
| Aplicación | Diez pods principales listos; deployments y StatefulSets listos; tres PVC `Bound` |
| Gateway API | Gateway `Programmed=True`, HTTPRoute aceptada, IP `203.0.113.20`; POST a `/grpc-202303204` respondió `202 Accepted` |
| Operator | CR `ValkeyInstance/reports` y `GrafanaInstance/dashboard` reconciliados, ambos workloads listos |
| Observabilidad | Prometheus respondió `military_reports_total=6574` y `military_country_reports_total{country="CHN"}=1314` tras la carga. Dashboard Grafana `mumnk8s-reports` con 12 paneles provisionados |

## Carga y escalado

Locust 2.31.8 ejecutó 50 usuarios durante 90 segundos contra el Gateway: **6569 POST, 0 fallos, 74,58 solicitudes/s**, media 353 ms, mediana 240 ms, p95 810 ms y p99 2900 ms. El HPA de Rust pasó de una a dos réplicas durante la carga y regresó a una después. La cola RabbitMQ terminó con cero mensajes pendientes y cero sin confirmar. Son valores de esa ejecución, no garantías de rendimiento futuras.

## Falco

Falco 0.45.1 quedó en un DaemonSet listo con driver modern eBPF. Su imagen, `falcoctl` y el plugin OCI de contenedores con su firma se sirvieron desde Zot. Una lectura de `/app/secretos/comprobacion` dentro de `rabbitmq-0` produjo la alerta personalizada `MUMNK8s intento de leer archivo sensible`, con proceso `cat`, archivo, contenedor, pod y namespace en el registro. El archivo de prueba fue eliminado. Véase [falco.md](falco.md) para repetir la prueba.

## Acceso durante la revisión

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
gcloud container clusters get-credentials mumnk8s-gke-c --zone us-central1-c --project ID_DEL_PROYECTO
kubectl get pods,deployments,statefulsets,hpa,pvc,gateway,httproute -n mumnk8s
kubectl get valkeyinstances,grafanainstances -n mumnk8s
kubectl get daemonset,pods -n falco
kubectl -n mumnk8s port-forward svc/grafana-dashboard 13000:3000
```

Con el túnel abierto, Grafana está en `http://localhost:13000/d/mumnk8s-reports/m-u-m-n-k8s-reportes-militares?orgId=1&refresh=10s`. Usuario `admin`; la clave se consulta localmente en WSL con `cat ~/.config/mumnk8s/grafana-password`. No publicar esa clave ni el kubeconfig.

## Límites y entrega

El presupuesto de USD 25 es una alerta, **no un límite de gasto**. VM, nodo GKE, discos y balanceador siguen consumiendo crédito mientras existan. El repositorio y este Markdown documentan el despliegue; faltan la entrega manual en UEDI y la defensa ante el auxiliar. Dapr es opcional y no se desplegó.
