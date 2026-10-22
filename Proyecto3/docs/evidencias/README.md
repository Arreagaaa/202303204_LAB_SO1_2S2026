# Capturas de funcionamiento

Capturas recibidas el 22 de octubre de 2026. Muestran el estado en el momento indicado; no sustituyen la demostración presencial.

| Archivo | Qué acredita |
|---|---|
| [zot-catalogo.png](zot-catalogo.png) | Catálogo de 16 repositorios publicado en Zot |
| [grafana-dashboard.png](grafana-dashboard.png) | Paneles de máximos, mínimos, modas, top países y métricas CHN |
| [grafana-locust.png](grafana-locust.png) | Grafana y Locust visibles durante la carga |
| [locust-estadisticas.png](locust-estadisticas.png) | POST a `/grpc-202303204` y cero fallos |
| [locust-graficas.png](locust-graficas.png) | Solicitudes/s, latencia y usuarios de Locust |
| [gke-nodo-pods.png](gke-nodo-pods.png) | Nodo GKE `Ready` y pods de aplicación `Running` |
| [gateway-hpa-operator.png](gateway-hpa-operator.png) | Gateway programado, HPA y recursos personalizados del operador |
| [valkey-datos.png](valkey-datos.png) | CRs de Valkey/Grafana y cinco claves en Valkey |
| [falco-api-rust.png](falco-api-rust.png) | Alerta real de Falco desde el pod de la API Rust |

El HPA aparece configurado en la captura, con una réplica después de la carga. Su aumento a dos réplicas durante la prueba anterior está registrado en [evidencia-gcp.md](../evidencia-gcp.md); esta captura no muestra el instante del escalamiento.
