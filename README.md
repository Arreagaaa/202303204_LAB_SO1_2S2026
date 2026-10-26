# Laboratorio de Sistemas Operativos 1 · 2S 2026

Tres proyectos que recorren virtualización, observabilidad del sistema Linux y despliegue de servicios en Kubernetes. Cada carpeta contiene el código, las instrucciones de ejecución y la documentación técnica de su proyecto.

| Proyecto | Qué construye | Tecnologías principales | Por dónde empezar |
|---|---|---|---|
| [Proyecto 1](Proyecto1/) | Tres API REST en Go distribuidas entre máquinas virtuales. Las imágenes se comparten mediante un registro Zot. | KVM, containerd, Podman, Docker, Go, Zot | [Guía de instalación](Proyecto1/docs/guia-instalacion.md) · [Arquitectura](Proyecto1/docs/manual-tecnico.md) |
| [Proyecto 2](Proyecto2/) | Un módulo de kernel expone métricas en `/proc`; un daemon decide qué contenedores retirar y confirma las señales mediante eBPF. | C, eBPF, Go, cron, Docker, Valkey, Grafana | [README del proyecto](Proyecto2/README.md) · [Guía de instalación](Proyecto2/docs/guia-instalacion.md) |
| [Proyecto 3](Proyecto3/) | Una plataforma de reportes con API en Rust, servicios Go, cola de mensajes, almacenamiento, métricas y escalado en Kubernetes. | Rust, Go, gRPC, RabbitMQ, Valkey, Helm Operator, Gateway API, HPA, Prometheus, Grafana, Falco, Locust | [README del proyecto](Proyecto3/README.md) · [Guía de instalación](Proyecto3/docs/guia-instalacion.md) |

## Recorrido sugerido

1. **Proyecto 1:** consulta el [manual técnico](Proyecto1/docs/manual-tecnico.md) para entender cómo se comunican las VM y cómo Zot distribuye las imágenes de las API.
2. **Proyecto 2:** revisa el [flujo del módulo, daemon y sonda eBPF](Proyecto2/docs/manual-tecnico.md). La ejecución funcional requiere Linux con permisos para cargar módulos y adjuntar eBPF.
3. **Proyecto 3:** sigue el [flujo de datos](Proyecto3/docs/manual-tecnico.md) desde la API hasta Grafana. Las [evidencias de GCP](Proyecto3/docs/evidencia-gcp.md) documentan una ejecución puntual; no presuponen que el despliegue siga activo.

Las direcciones de infraestructura de los ejemplos son ilustrativas. Las pruebas locales generan sus contraseñas fuera de Git; las guías de cada proyecto explican cómo configurar el entorno propio.
