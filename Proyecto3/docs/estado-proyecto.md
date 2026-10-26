# Estado de Proyecto 3 al 22 de octubre de 2026

**Estudiante:** Christian Javier Rivas Arreaga, carné 202303204. Ruta pública `/grpc-202303204`; país asignado **CHN**.

## Estado verificado

El sistema se desplegó en un proyecto GCP privado. GKE tuvo un nodo listo, Zot HTTPS sirvió las imágenes y el Gateway API respondió públicamente. La API Rust, los servicios Go, RabbitMQ, consumer, Valkey, Prometheus, Grafana y el operador estuvieron listos. Los CR `ValkeyInstance` y `GrafanaInstance` fueron reconciliados. [Evidencia detallada](evidencia-gcp.md).

- Locust en GKE: 50 usuarios, 90 segundos, 6569 POST, cero fallos; HPA de Rust pasó de una a dos réplicas.
- Comparación corta de servicios Go: con 20 usuarios y 45 segundos, una y dos réplicas dieron ~47,9 solicitudes/s sin fallos; dos réplicas no mejoraron el p95. Se restauró una réplica por deployment.
- Prometheus registró 6574 reportes totales y 1314 del país CHN tras la carga. Grafana tiene doce paneles provisionados.
- Falco generó una alerta real desde `api-rust` al leer `/proc/1/environ` sin mostrar su contenido (22 de octubre).
- La comparación de Valkey con primario solo y primario más réplica de datos se ejecutó localmente con 1000 escrituras por escenario; el PVC activo de GKE no se modificó.
- Operator SDK v1.37.0 generó `PROJECT`, `Makefile` y `config/` para ambos CR mediante `init` y `create api`; el runtime desplegado conserva los charts y manifiestos comprobados.
- La validación local anterior incluyó Go, Rust, Helm, Docker Compose, kind, reconciliación y persistencia del operador. Los resultados locales y de GKE deben interpretarse por separado.

## Pendiente para la entrega

1. Entregar el repositorio privado y el manual Markdown en UEDI antes del plazo indicado por la cátedra. Confirmar el acceso de los revisores exigidos por el enunciado.
2. Preparar la demostración y defensa oral. El funcionamiento observado no sustituye la evaluación del auxiliar.
3. Revisar el consumo de GCP cada día y eliminar recursos cuando termine la evaluación. La alerta mensual de USD 25 no detiene los cargos.
4. Llevar abiertas las capturas de [evidencias/](evidencias/README.md) y las pantallas del sistema para la evaluación presencial.
5. Consultar al auxiliar si la comparación local de Valkey ya documentada satisface el criterio o si exige además una demostración en GKE. La prueba cloud tendría que ser aislada del Valkey activo.

Dapr era opcional y no se implementó. El enunciado menciona Gateway API explícitamente; se usó Gateway API aunque una tabla secundaria conserve la palabra Ingress.
