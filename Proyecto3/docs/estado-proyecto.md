# Estado de Proyecto 3 — 19 de octubre de 2026

**Estudiante:** Christian Javier Rivas Arreaga, **carné 202303204**. La ruta principal correcta es `/grpc-202303204`. Su último dígito es 4; por el enunciado, el país asignado del dashboard es **CHN**.

## Avance estimado

- **Preparación del código y manifiestos antes de GCP: 92 %**. Estimación de ingeniería: los ocho contenedores construyen y el flujo pasó en Docker Compose y Kubernetes local. Queda validar los detalles propios de GKE y Zot.
- **Cumplimiento demostrable de la rúbrica completa: 68 %**. Estimación conservadora de evidencia disponible; **no es nota oficial**. El proyecto requiere GKE funcional y entrega en UEDI para calificarse.

| Criterio de la rúbrica | Máximo | Evidencia disponible | Estimación |
|---|---:|---|---:|
| Gateway API en GKE | 5 | HTTPRoute probada con Envoy Gateway en kind; GKE pendiente | 3 |
| API Rust y HPA | 5 | API y HPA probados en kind; escaló de 1 a 3 réplicas | 5 |
| Servicios Go | 10 | Dos contenedores por pod, protobuf gRPC y flujo real en Compose | 9 |
| Zot en VM GCP | 10 | Guía y referencias de imágenes; VM/registry pendientes | 1 |
| RabbitMQ y consumer | 15 | Publicación y consumo real en Compose y kind | 12 |
| Valkey con Operator | 20 | Creación, actualización, eliminación y persistencia probadas en kind | 16 |
| Grafana | 25 | Exporter y Prometheus con datos reales; dashboard provisionado en kind | 18 |
| Locust, HPA y comparación 1/2 réplicas | 5 | Carga local y escalado HPA probados; comparación controlada pendiente | 3 |
| Manual técnico | 2 | Markdown escrito; faltan evidencias y conclusiones reales | 1 |
| Preguntas en calificación | 3 | No evaluadas | 0 |
| **Total estimado** | **100** | | **68** |

## Qué se corrigió

1. Se reemplazó el carné incompleto en la ruta de Rust, Gateway, Locust y documentación.
2. Se generó el código Go desde `report.proto` y se sustituyó el codec JSON gRPC por protobuf estándar.
3. Se separaron los dos deployments Go en dos contenedores funcionales cada uno: REST + cliente gRPC y servidor gRPC + publicador RabbitMQ.
4. Se añadió confirmación de publicación y mensajes persistentes en RabbitMQ; el consumer comprueba la respuesta de Valkey antes de confirmar el mensaje.
5. Se añadió el Helm Operator con watches, RBAC y deployment. Valkey ya no se despliega por un StatefulSet paralelo; solo lo administra el CR.
6. Se añadió un exporter desde Valkey, Prometheus y dashboard con los doce indicadores, incluido CHN.
7. Se prepararon imágenes de todos los componentes para consumirse desde Zot y se reparó el Dockerfile del consumer.
8. Se añadió `serde_json`, se fijó un `Cargo.lock` y se probó la API Rust.

## Verificaciones locales realizadas

- `go test` y `go vet` de todos los módulos: correctos.
- `cargo test` de Rust: 4/4 pruebas correctas.
- `helm lint` y `helm template` de ambos charts: correctos.
- YAML: 12 archivos analizados, 29 recursos de Kubernetes, sin identidad duplicada.
- Dashboard JSON: 12 paneles y JSON válido.
- Docker en WSL Ubuntu: construidas las siete imágenes de aplicación y la del operador.
- Docker Compose: API → servicios Go → RabbitMQ → consumer → Valkey; reporte CHN guardado y visible en métricas.
- Prometheus listo; Grafana respondió correctamente y mostró el dashboard provisionado.
- Locust local, 10 usuarios durante 1 minuto: **1857 solicitudes, 0 fallos**, mediana de 8 ms y percentil 95 de 14 ms. Son mediciones locales, no de GKE.
- kind en WSL: pods principales `Ready`, reporte CHN recibido por la API y por Envoy Gateway; registro confirmado en Valkey.
- Helm Operator: Valkey y Grafana creados por CR. El cambio de `maxMemory` actualizó el StatefulSet y conservó el reporte en su PVC; una instancia adicional se creó y eliminó correctamente.
- Metrics Server y HPA: durante la carga, la API Rust escaló de 1 a 3 réplicas.
- Locust a través de Envoy Gateway en kind: **10873 solicitudes, 0 fallos** durante 90 segundos con 100 usuarios; mediana 480 ms y percentil 95 660 ms. La cola RabbitMQ se drenó después de la carga.

Las pruebas revelaron y corrigieron un flag inválido del operator, una colisión de nombres de releases Helm y la ausencia de subrecursos `status` en ambos CRD. La primera carga de Compose detectó un problema de arranque del publicador y se añadió reinicio automático. En kind, Envoy Gateway enruta correctamente por `port-forward`; su condición `Programmed` queda en falso porque kind no asigna una IP externa al Service tipo LoadBalancer.

## Revisión de seguridad antes del commit

- Gitleaks v8.30.1 examinó el directorio completo de Proyecto3 y sus seis commits anteriores: sin hallazgos.
- No hay archivos `.env`, claves privadas, certificados ni kubeconfigs entre los archivos del proyecto que entrarían al commit. `.gitignore` excluye estos formatos y credenciales de GCP.
- Los manifiestos de producción ya no incluyen contraseñas. RabbitMQ requiere el Secret externo `rabbitmq-credentials` y Grafana el Secret externo `grafana-admin`; ambos se crearán fuera de Git durante el despliegue.
- `compose.local.yaml` y `k8s/local/` contienen solamente la contraseña desechable `x`, destinada a pruebas locales. No debe usarse en GCP.

## Pendientes para cerrar el proyecto

1. Probar la GatewayClass de GKE y Zot en GCP. La prueba de Gateway, HPA y reconciliación en kind ya pasó.
2. En GCP, crear GKE con nodos N1 y Zot HTTPS en una VM N1; publicar y descargar todas las imágenes desde Zot, más el OCI artifact documentado.
3. Repetir en GKE la creación, actualización y eliminación de CRs, así como persistencia y reinicio de pods.
   El runtime Helm Operator se empaquetó manualmente siguiendo el ejemplo de Semana 11; no se ejecutó `operator-sdk init/create api` en este equipo. Confirmar con la cátedra si exigirá los archivos generados por ese CLI además del funcionamiento.
4. Comprobar visualmente los doce paneles de Grafana con datos reales. El exporter recorre listas completas de Valkey en cada scrape; evaluar su costo bajo la carga objetivo.
5. Medir Locust, HPA y el rendimiento con 1 y 2 réplicas en GKE; documentar resultados y capturas.
6. Añadir al manual capturas funcionales, resultados, conclusiones y, si se hace Falco, alerta real. Dapr es opcional y sigue pendiente.
7. Verificar repositorio privado, acceso de @JoseLorenzana272 y @KINGR0X, entrega en UEDI y manual Markdown. La cuenta de GCP y el repositorio para esta fase aún no han sido preparados por el estudiante.

**Nota del enunciado:** su descripción exige Gateway API, aunque una tabla de detalle conserva una mención contradictoria a Ingress Controller. Se siguió la instrucción explícita de Gateway API.
