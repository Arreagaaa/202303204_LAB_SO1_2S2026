# Despliegue de Proyecto 3 en GCP

## Recursos

- Proyecto: `ID_DEL_PROYECTO`
- Registro Zot: VM `mumnk8s-zot`, N1 standard 1, `us-central1-a`
- Registro HTTPS: `203-0-113-10.sslip.io`
- Clúster GKE Standard: `mumnk8s-gke-c`, N1 standard 2, un nodo zonal en `us-central1-c`
- Presupuesto: alerta mensual de USD 25 para este proyecto. La alerta **no limita ni detiene** el gasto.

Las credenciales se guardan en WSL bajo `~/.config/mumnk8s/` con permisos restringidos. No se deben copiar a Git ni a capturas de pantalla.

## Crear el ambiente desde cero

Esta guía recoge los pasos usados para desplegar el sistema y permite repetirlos en otro proyecto de GCP. Los comandos se ejecutan en WSL Ubuntu desde la carpeta `Proyecto3`. El proyecto necesita una cuenta de facturación activa y permisos para crear máquinas virtuales, clústeres y reglas de red. Varios proyectos vinculados a una misma cuenta de facturación comparten sus créditos.

Antes de abrir WSL, crea el proyecto desde el selector de proyectos de la consola de GCP con **Proyecto nuevo**. Anota el ID que asigna Google, selecciona ese proyecto y vincúlalo a la cuenta de facturación que usarán. En **Facturación > Presupuestos y alertas**, crea un presupuesto filtrado por el proyecto y agrega avisos por correo. Comprueba el proyecto seleccionado antes de ejecutar cualquier comando: crear recursos en el proyecto equivocado también consume crédito.

Instalar y autenticar la CLI oficial de Google Cloud, y tener disponibles `gcloud`, `kubectl`, `gke-gcloud-auth-plugin`, Docker, Helm, ORAS, Python 3 y OpenSSL. [Google explica cómo instalar la CLI](https://cloud.google.com/sdk/docs/install) y el [plugin de autenticación de GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl). Antes de crear recursos, comprobar la identidad y el proyecto:

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
gcloud auth login
export GCP_PROJECT_ID=ID_REAL_DEL_PROYECTO
gcloud config set project "$GCP_PROJECT_ID"
gcloud auth list
gcloud billing projects describe "$GCP_PROJECT_ID"
gcloud services enable compute.googleapis.com container.googleapis.com --project "$GCP_PROJECT_ID"
```

La salida de `gcloud billing projects describe` debe indicar `billingEnabled: true`. El propietario de la facturación puede crear en la consola un presupuesto mensual filtrado por este proyecto y configurar avisos por correo. Un presupuesto de **solo alertas** no detiene recursos ni gastos. Revisar el informe de costos cada día; los cargos pueden aparecer con retraso.

Elegir una zona con disponibilidad de máquinas N1. En esta instalación funcionaron `us-central1-a` para Zot y `us-central1-c` para GKE. Reservar la IP del registro y abrir solo HTTP/HTTPS para el certificado y el registry:

```bash
export GCP_REGION=us-central1
export ZOT_ZONE=us-central1-a
export GKE_ZONE=us-central1-c
gcloud compute addresses create mumnk8s-zot-ip --region "$GCP_REGION" --project "$GCP_PROJECT_ID"
export ZOT_IP=$(gcloud compute addresses describe mumnk8s-zot-ip \
  --region "$GCP_REGION" --project "$GCP_PROJECT_ID" --format='value(address)')
export ZOT_HOST="${ZOT_IP//./-}.sslip.io"
gcloud compute firewall-rules create mumnk8s-zot-https \
  --network=default --allow=tcp:80,tcp:443 --source-ranges=0.0.0.0/0 \
  --target-tags=mumnk8s-zot --project "$GCP_PROJECT_ID"
```

El nombre de Zot sale de la IP reservada. En esta instalación, `203.0.113.10` se convirtió en `203-0-113-10.sslip.io`. El servicio DNS de `sslip.io` devuelve esa misma IP cuando se consulta el nombre. No se compró un dominio ni se creó una zona de Cloud DNS. Caddy recibe las conexiones por HTTPS en la VM y las envía a Zot, que escucha en `127.0.0.1:5000`. Caddy obtiene y renueva el certificado TLS. El nombre depende de `sslip.io`, mientras que la VM, la IP, el disco y el tráfico pertenecen a la infraestructura de GCP.

La red `default` de GCP puede incluir una regla que permita SSH desde cualquier IP. Revisarla y limitarla según la red desde la que se administrará la VM. La VM de Zot no necesita una cuenta de servicio de Google:

```bash
gcloud compute instances create mumnk8s-zot \
  --zone "$ZOT_ZONE" --machine-type=n1-standard-1 \
  --image-family=ubuntu-2404-lts-amd64 --image-project=ubuntu-os-cloud \
  --boot-disk-size=25GB --boot-disk-type=pd-standard \
  --address="$ZOT_IP" --tags=mumnk8s-zot \
  --no-service-account --no-scopes --project "$GCP_PROJECT_ID"
```

Crear una contraseña distinta para este registry y conservarla fuera del repositorio. El script de la VM elimina su copia temporal al convertirla en un hash para Zot:

```bash
mkdir -p "$HOME/.config/mumnk8s"
chmod 700 "$HOME/.config/mumnk8s"
openssl rand -hex 24 > "$HOME/.config/mumnk8s/registry-password"
chmod 600 "$HOME/.config/mumnk8s/registry-password"
gcloud compute ssh mumnk8s-zot --zone "$ZOT_ZONE" --project "$GCP_PROJECT_ID" \
  --command='mkdir -p ~/zot-setup'
gcloud compute scp scripts/gcp/setup-zot-vm.sh scripts/gcp/zot-config.json \
  scripts/gcp/Caddyfile.template mumnk8s-zot:~/zot-setup/ \
  --zone "$ZOT_ZONE" --project "$GCP_PROJECT_ID"
gcloud compute scp "$HOME/.config/mumnk8s/registry-password" \
  mumnk8s-zot:~/registry-password --zone "$ZOT_ZONE" --project "$GCP_PROJECT_ID"
gcloud compute ssh mumnk8s-zot --zone "$ZOT_ZONE" --project "$GCP_PROJECT_ID" \
  --command="bash ~/zot-setup/setup-zot-vm.sh $ZOT_HOST"
curl -I "https://$ZOT_HOST/v2/"
```

Una respuesta HTTP `401` de `/v2/` confirma que HTTPS responde y pide autenticación. Si Caddy aún no obtuvo el certificado, esperar y comprobar los logs de `mumnk8s-caddy` en la VM. El nombre `sslip.io` depende de un servicio DNS externo; si no resuelve, usar un dominio propio apuntado a la IP reservada.

Crear el clúster Standard con un nodo N1 y Gateway API. El nodo necesita una IP de salida para descargar imágenes desde Zot:

```bash
gcloud container clusters create mumnk8s-gke-c \
  --zone "$GKE_ZONE" --num-nodes=1 --machine-type=n1-standard-2 \
  --disk-size=50 --disk-type=pd-standard --release-channel=regular \
  --gateway-api=standard --enable-ip-alias --project "$GCP_PROJECT_ID"
gcloud container clusters get-credentials mumnk8s-gke-c \
  --zone "$GKE_ZONE" --project "$GCP_PROJECT_ID"
kubectl get nodes
kubectl get gatewayclass
```

Si GCP responde que no hay capacidad N1 en la zona, elegir otra antes de seguir y actualizar `GKE_ZONE`. Confirmar que no quedó un clúster incompleto con `gcloud container clusters list`. A partir de aquí, publicar imágenes, desplegar la aplicación y Falco con los comandos de las secciones siguientes. Al repetir el proceso en otro proyecto, establecer el `GCP_PROJECT_ID`, la IP y el host de Zot correspondientes. Los scripts de publicación y Falco reciben el host como argumento; el de GKE lee `GCP_PROJECT_ID` del entorno.

## Publicar de nuevo

En WSL, desde `Proyecto3`:

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
bash scripts/gcp/publish-images.sh "$ZOT_HOST"
```

El script construye las siete imágenes de aplicación y la del operador con las referencias reales de Zot en sus charts. También publica RabbitMQ, Valkey, Grafana, Prometheus, Falco y falcoctl en Zot, además del archivo `k8s/00-namespace.yaml` como artefacto OCI y el plugin firmado de Falco.

## Desplegar de nuevo

```bash
export PATH="$HOME/google-cloud-sdk/bin:$HOME/.local/bin:$PATH"
bash scripts/gcp/deploy-gke.sh "$ZOT_HOST" "$GKE_ZONE" mumnk8s-gke-c
```

El script obtiene las credenciales del clúster y comprueba el contexto antes de aplicar recursos. Crea los Secret de registro, RabbitMQ y Grafana a partir de archivos locales fuera del repositorio.

Falco se instala después de publicar las imágenes y el plugin firmado:

```bash
bash scripts/gcp/deploy-falco.sh "$ZOT_HOST"
```

## Comprobaciones

```bash
kubectl get nodes
kubectl get pods,deployments,statefulsets,hpa,pvc,gateway,httproute -n mumnk8s
kubectl get valkeyinstances,grafanainstances -n mumnk8s
oras repo ls "$ZOT_HOST"
kubectl get daemonset,pods -n falco
```

Para ver Grafana desde Windows, iniciar este túnel en WSL y dejar la terminal abierta:

```bash
kubectl -n mumnk8s port-forward svc/grafana-dashboard 13000:3000
```

Abrir `http://localhost:13000/d/mumnk8s-reports/m-u-m-n-k8s-reportes-militares?orgId=1&refresh=10s`. El usuario es `admin`; la contraseña está en `~/.config/mumnk8s/grafana-password` dentro de WSL.

Los resultados medidos de Gateway, Locust, HPA, Prometheus y Falco están en [evidencia-gcp.md](evidencia-gcp.md). Para repetir Falco, seguir [falco.md](falco.md).

Para la entrega hacen falta capturas del clúster y sus pods listos, la ruta Gateway, las imágenes en Zot, la ejecución de Locust y el dashboard de Grafana. Antes de la defensa se deben comprobar los doce paneles con datos del proyecto.

## Costo y cierre

La VM, el nodo GKE, los discos y el balanceador del Gateway consumen crédito mientras existan. Revisa la facturación del proyecto a diario. Antes de eliminar recursos, guarda las capturas y los datos que necesites: al borrar el clúster o la VM también puedes perder sus discos.

El presupuesto configurado es **solo una alerta**, no un tope. La cuenta de facturación está administrada por su propietario; cualquier cambio de facturación se coordina con él. La eliminación de recursos se hace únicamente después de guardar la evidencia y confirmar que ya no se necesitan.

Después de la calificación, guarda las capturas y confirma el proyecto activo. Estos comandos eliminan el clúster, la VM, la IP reservada y la regla HTTPS de **ese proyecto**. No los ejecutes mientras necesites mostrar el sistema:

```bash
gcloud config set project "$GCP_PROJECT_ID"
gcloud container clusters delete mumnk8s-gke-c \
  --zone "$GKE_ZONE" --project "$GCP_PROJECT_ID"
gcloud compute instances delete mumnk8s-zot \
  --zone "$ZOT_ZONE" --project "$GCP_PROJECT_ID"
gcloud compute addresses delete mumnk8s-zot-ip \
  --region "$GCP_REGION" --project "$GCP_PROJECT_ID"
gcloud compute firewall-rules delete mumnk8s-zot-https \
  --project "$GCP_PROJECT_ID"
```

Revisa después **Compute Engine > Discos** y **Red > Balanceo de cargas** en la consola. Si queda algún recurso, comprueba a qué proyecto pertenece antes de eliminarlo. El informe de facturación puede tardar en reflejar el cierre.
