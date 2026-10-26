# Preparación para publicar el repositorio

Los ejemplos de infraestructura usan direcciones de documentación. Las contraseñas de las pruebas locales se generan fuera de Git: `Proyecto2/infra/.env.local` y `~/.config/mumnk8s-local/` para Proyecto 3. Las credenciales de GCP y Zot se administran fuera de este repositorio.

Antes de cambiar la visibilidad del repositorio:

1. Verifica que ninguna contraseña de pruebas anteriores siga activa. Rota las cuentas locales de RabbitMQ y Grafana si conservas sus volúmenes.
2. Revisa los archivos nuevos y las capturas antes de hacer commit. No incluyas kubeconfig, tokens, claves privadas ni archivos `.env`.
3. Ejecuta un detector de secretos sobre los archivos y todos los commits, por ejemplo `gitleaks dir .` y `gitleaks git .`.
4. Comprueba las ramas y etiquetas remotas: una rama antigua puede conservar objetos que ya no estén en `main`.
5. Si se encuentra una credencial real, revócala en el proveedor. Borrar el archivo o reescribir Git no sustituye la revocación.

El historial de Git conserva metadatos de autor y el proyecto conserva identificadores académicos deliberados. Revísalos antes de publicar si también deseas anonimato personal.
