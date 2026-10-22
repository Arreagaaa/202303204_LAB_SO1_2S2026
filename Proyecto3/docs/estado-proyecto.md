# Estado de Proyecto 3 — 21 de octubre de 2026

**Estudiante:** Christian Javier Rivas Arreaga, carné 202303204. Ruta pública `/grpc-202303204`; país asignado **CHN**.

## Estado verificado

El sistema está desplegado en el proyecto GCP `ID_DEL_PROYECTO`. GKE tiene un nodo listo, Zot HTTPS sirve las imágenes y el Gateway API responde públicamente. La API Rust, los servicios Go, RabbitMQ, consumer, Valkey, Prometheus, Grafana y el operador están listos. Los CR `ValkeyInstance` y `GrafanaInstance` fueron reconciliados. [Evidencia detallada](evidencia-gcp.md).

- Locust en GKE: 50 usuarios, 90 segundos, 6569 POST, cero fallos; HPA de Rust pasó de una a dos réplicas.
- Prometheus registró 6574 reportes totales y 1314 del país CHN tras la carga. Grafana tiene doce paneles provisionados.
- Falco generó una alerta real al leer un archivo bajo `/app/secretos` desde un pod de la aplicación.
- Operator SDK v1.37.0 generó `PROJECT`, `Makefile` y `config/` para ambos CR mediante `init` y `create api`; el runtime desplegado conserva los charts y manifiestos comprobados.
- La validación local anterior incluyó Go, Rust, Helm, Docker Compose, kind, reconciliación y persistencia del operador. Los resultados locales y de GKE deben interpretarse por separado.

## Pendiente para la entrega

1. Entregar el repositorio privado y el manual Markdown en UEDI antes del plazo indicado por la cátedra. Confirmar el acceso de los revisores exigidos por el enunciado.
2. Preparar la demostración y defensa oral. El funcionamiento observado no sustituye la evaluación del auxiliar.
3. Revisar el consumo de GCP cada día y eliminar recursos cuando termine la evaluación. La alerta mensual de USD 25 no detiene los cargos.

Dapr era opcional y no se implementó. El enunciado menciona Gateway API explícitamente; se usó Gateway API aunque una tabla secundaria conserve la palabra Ingress.
