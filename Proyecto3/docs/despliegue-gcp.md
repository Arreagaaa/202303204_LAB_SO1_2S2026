# Despliegue de Proyecto 3 en GCP

## Recursos

- Proyecto: `ID_DEL_PROYECTO`
- Registro Zot: VM `mumnk8s-zot`, N1 standard 1, `us-central1-a`
- Registro HTTPS: `203-0-113-10.sslip.io`
- Clúster GKE Standard: `mumnk8s-gke-c`, N1 standard 2, un nodo zonal en `us-central1-c`
- Presupuesto: alerta mensual de USD 25 para este proyecto. La alerta **no limita ni detiene** el gasto.

Las credenciales se guardan en WSL bajo `~/.config/mumnk8s/` con permisos restringidos. No se deben copiar a Git ni a capturas de pantalla.

## Publicar de nuevo

En WSL, desde `Proyecto3`:

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
bash scripts/gcp/publish-images.sh 203-0-113-10.sslip.io
```

El script construye las siete imágenes de aplicación y la del operador con las referencias reales de Zot en sus charts. También publica RabbitMQ, Valkey, Grafana, Prometheus, Falco y falcoctl en Zot, además del archivo `k8s/00-namespace.yaml` como artefacto OCI y el plugin firmado de Falco.

## Desplegar de nuevo

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
bash scripts/gcp/deploy-gke.sh 203-0-113-10.sslip.io us-central1-c mumnk8s-gke-c
```

El script obtiene las credenciales del clúster y comprueba el contexto antes de aplicar recursos. Crea los Secret de registro, RabbitMQ y Grafana a partir de archivos locales fuera del repositorio.

Falco se instala después de publicar las imágenes y el plugin firmado:

```bash
bash scripts/gcp/deploy-falco.sh 203-0-113-10.sslip.io
```

## Comprobaciones

```bash
kubectl get nodes
kubectl get pods,deployments,statefulsets,hpa,pvc,gateway,httproute -n mumnk8s
kubectl get valkeyinstances,grafanainstances -n mumnk8s
oras repo ls 203-0-113-10.sslip.io
kubectl get daemonset,pods -n falco
```

Para ver Grafana desde Windows, iniciar este túnel en WSL y dejar la terminal abierta:

```bash
kubectl -n mumnk8s port-forward svc/grafana-dashboard 13000:3000
```

Abrir `http://localhost:13000/d/mumnk8s-reports/m-u-m-n-k8s-reportes-militares?orgId=1&refresh=10s`. El usuario es `admin`; la contraseña está en `~/.config/mumnk8s/grafana-password` dentro de WSL.

Los resultados medidos de Gateway, Locust, HPA, Prometheus y Falco están en [evidencia-gcp.md](evidencia-gcp.md). Para repetir Falco, seguir [falco.md](falco.md).

## Costo y cierre

La VM, el nodo GKE, los discos y el balanceador del Gateway consumen crédito mientras existan. Revisar la facturación del proyecto a diario. Cuando termine la evaluación, detener el consumo eliminando los recursos del proyecto con los comandos apropiados; conservar antes cualquier evidencia o dato que se necesite. La eliminación de clúster o VM destruye los datos almacenados en sus discos si se usa la configuración normal de borrado.

El presupuesto configurado es **solo una alerta**, no un tope. La cuenta de facturación pertenece al compañero que creó el proyecto. Coordinar con él antes de borrar recursos o modificar facturación.
