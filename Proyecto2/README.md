# Proyecto 2 - Sonda de kernel y telemetría de contenedores

**Curso:** Sistemas Operativos 1  
**Carnet:** 202303204

Este proyecto reúne un módulo de kernel en C, una sonda eBPF y un daemon en Go. El módulo publica un snapshot de memoria y procesos en `/proc`; el daemon lo lee y decide qué procesos candidatos puede retirar sin romper las cantidades mínimas del laboratorio.

## Estado actual

- `kernel-module/`: módulo C con `seq_file`, `task_struct` y `Makefile` para Ubuntu.
- `daemon/`: parser del formato `/proc`, modelo de datos y política de selección.
- `cronjob/`: creación periódica de cinco contenedores y un intruso.
- `docker-images/`: imágenes de alto consumo de RAM, alto consumo de CPU, bajo consumo y lectura de prueba.
- `ebpf-probe/`: programa eBPF que publica eventos de `sys_enter_kill` en un ring buffer.
- `infra/`: Valkey y Grafana con configuración versionada.
- `docs/`: guía de instalación, manual técnico y límites de las pruebas desde Windows.

La compilación del módulo, la carga eBPF, cron y Docker requieren una máquina Linux. Desde Windows se puede revisar el código, preparar las imágenes y editar la configuración, pero no se deben presentar como pruebas funcionales del kernel.

## Contrato del archivo `/proc`

El módulo genera estas líneas de memoria:

```text
MEM_TOTAL_KB:...
MEM_FREE_KB:...
MEM_USED_KB:...
```

Después escribe una cabecera y una fila por proceso:

```text
PID,NAME,CMD_OR_CONTAINER_ID,VSZ_KB,RSS_KB,MEM_PCT,CPU_PCT,CLASS
```

En la primera versión, `CLASS` sale como `unknown`. La identificación final de contenedores debe cruzar el PID con Docker y sus etiquetas desde el daemon; no conviene inventarla dentro del módulo.

## Ejecución en Ubuntu

```bash
cd Proyecto2
chmod +x kernel-module/load_module.sh cronjob/spawn_containers.sh
make -C kernel-module
./kernel-module/load_module.sh load
cd daemon && go run . -once
```

La guía completa está en [docs/guia-instalacion.md](docs/guia-instalacion.md).
