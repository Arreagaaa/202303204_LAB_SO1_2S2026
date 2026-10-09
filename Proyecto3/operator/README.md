# Operator Helm-based

## Proposito

El Operator permite administrar Valkey y Grafana mediante recursos personalizados de Kubernetes. El usuario declara el estado deseado en un Custom Resource y el reconciliador Helm crea o actualiza los recursos asociados.

## Recursos administrados

### ValkeyInstance

Representa una instancia de Valkey y permite definir:

- Numero de replicas.
- Tamano de almacenamiento.

El chart crea un StatefulSet, un Service y un PVC.

### GrafanaInstance

Representa una instancia de Grafana y permite definir:

- Puerto del Service.
- Service de Valkey utilizado como datasource.

El chart crea un Deployment, un Service, un Secret y el provisioning inicial de Grafana.

## Crear el proyecto con Operator SDK

Ejecutar desde una carpeta destinada exclusivamente al operador:

```bash
operator-sdk init --plugins helm
operator-sdk create api \
  --group mumnk8s \
  --version v1alpha1 \
  --kind ValkeyInstance
operator-sdk create api \
  --group mumnk8s \
  --version v1alpha1 \
  --kind GrafanaInstance
```

Copiar los charts de esta carpeta a la carpeta `helm-charts/` del proyecto generado y ajustar los valores del chart a los campos definidos en los CRDs.

## Instalar

Construir y publicar la imagen del operador:

```bash
make docker-build docker-push IMG="$ZOT/mumnk8s/operator:dev"
```

Instalar los manifests del operador:

```bash
make deploy IMG="$ZOT/mumnk8s/operator:dev"
```

Aplicar los recursos del proyecto:

```bash
kubectl apply -f ../k8s/07-crds.yaml
kubectl apply -f ../k8s/06-custom-resources.yaml
```

## Verificar la reconciliacion

```bash
kubectl get valkeyinstances,grafanainstances -n mumnk8s
kubectl get statefulsets,deployments,svc,pvc -n mumnk8s
kubectl logs -n mumnk8s deployment/mumnk8s-controller-manager -c manager
```

La instalacion se considera correcta cuando los Custom Resources existen, el operador no presenta errores y los recursos derivados llegan al estado `Ready`.
