# Manual técnico - Proyecto 2

**Curso:** Sistemas Operativos 1  
**Carnet:** 202303204

## 1. Propósito

El sistema observa procesos desde el kernel y deja la decisión operativa en un daemon de Go. Esa separación es intencional: el módulo no detiene contenedores y el daemon no intenta reconstruir las estructuras internas del kernel.

## 2. Flujo de datos

1. `continfo.ko` consulta la memoria del host y recorre `task_struct`.
2. El módulo publica un snapshot en `/proc/continfo_pr2_so1_202303204`.
3. El parser de Go convierte cada fila en un `model.Process`.
4. La política ordena los candidatos por RSS, VSZ, CPU y memoria.
5. Antes de retirar un contenedor, el daemon envía `SIGTERM` directamente a su PID mediante `syscall.Kill`.
6. La sonda eBPF captura `sys_enter_kill`; el daemon espera la coincidencia del PID en `trace_pipe`.
7. Solo después de esa confirmación se guarda el evento de eliminación en Valkey.
8. Después de detenerlo, el daemon ejecuta `docker rm` para eliminar el contenedor.
9. Grafana consulta las claves publicadas en Valkey.

## 2.1. Diagrama de arquitectura

El flujo completo de generación, análisis, eliminación y confirmación se representa en el siguiente diagrama:

![Arquitectura del sistema](images/07-arquitectura.png)

El diagrama cubre el ciclo solicitado: cronjob, contenedores, módulo `/proc`, daemon Go, decisión por política, syscall `sys_kill`, confirmación eBPF, Valkey y Grafana.

## 3. Módulo de kernel

El archivo `kernel-module/continfo.c` usa `seq_file`, que permite entregar el listado por partes cuando el archivo `/proc` se lee. La memoria está expresada en KB. El porcentaje de memoria se calcula con la memoria residente del proceso respecto de la RAM total. El porcentaje de CPU usa el tiempo acumulado del proceso y el tiempo transcurrido desde su inicio; por eso puede diferir de herramientas como `top`.

El campo `CMD_OR_CONTAINER_ID` utiliza temporalmente `task->comm`. La relación exacta entre un PID y un contenedor depende del runtime y debe resolverse desde el daemon mediante Docker y cgroups. El campo `CLASS` queda como `unknown` en el módulo; las etiquetas Docker son la fuente adecuada para distinguir `low`, `high`, `grafana` e `intruder`.

Las imágenes de prueba representan perfiles distintos: `go-client` reserva y toca memoria para producir consumo RSS sostenido, `alpine-highcpu` mantiene un ciclo de CPU intensivo y `alpine-lowconsumo` ejecuta `sleep 240`.

## 4. Política de decisiones

`internal/decision` recibe una fotografía completa y aplica estos límites:

- conserva al menos tres procesos clasificados como `low`;
- conserva al menos dos procesos clasificados como `high`;
- nunca selecciona un proceso clasificado como `grafana`;
- considera candidato un proceso que supere el límite de memoria o CPU;
- ordena primero por RSS, luego VSZ, CPU y porcentaje de memoria.

La política no ejecuta ningún `kill`. Esa operación pertenece al administrador de contenedores y queda pendiente de confirmación eBPF.

`internal/containerctl` resuelve el PID de cada contenedor mediante `docker inspect` y `internal/cronmanager` administra una entrada marcada en el crontab del usuario. Las dos piezas están aisladas para poder sustituir el comando Docker por el SDK oficial durante la integración final sin cambiar la política.

El cronjob garantiza tres contenedores `low` y dos `high` por ejecución. Los dos perfiles `high` se seleccionan aleatoriamente entre `go-client` y `alpine-highcpu`; después se mezclan las cinco posiciones para conservar el despliegue aleatorio sin violar los mínimos.

## 5. Sonda eBPF

`ebpf-probe/kill_tracer.bpf.c` se conecta al tracepoint `sys_enter_kill` y publica `sender_pid`, `target_pid` y `signal`. Además emite una línea con `bpf_printk` usando el prefijo `SO1_PR2_KILL` para que el daemon la lea desde `trace_pipe`. `vmlinux.h` se genera en la VM desde el BTF del kernel. El objeto eBPF no se puede compilar en Windows.

El listener de user space del daemon lee `trace_pipe`, filtra las líneas que contienen el prefijo `SO1_PR2_KILL` y asocia el PID del contenedor con la confirmación real de terminación.

## 6. Valkey y Grafana

Compose levanta Valkey en el puerto `6379` y Grafana en `3001`. Se dejaron las claves usadas por el dashboard:

- `system:ram:current`
- `system:ram:history`
- `containers:killed:count`
- `containers:killed:history`
- `containers:top:ram` (HASH con PID, nombre, ID y RSS)
- `containers:top:cpu` (HASH con PID, nombre, ID y CPU)
- `ebpf:events:count`
- `ebpf:events:raw`

El daemon publica estas claves desde su ciclo principal después de integrar el cliente de Valkey y el listener eBPF.
