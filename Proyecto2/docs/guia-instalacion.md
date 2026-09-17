# Guía de instalación - Proyecto 2

## 1. Requisitos

Las pruebas reales requieren Ubuntu con:

```bash
sudo apt update
sudo apt install -y build-essential linux-headers-$(uname -r) clang llvm libbpf-dev bpftool golang docker.io docker-compose-v2 cron
sudo usermod -aG docker "$USER"
```

Cierra la sesión y vuelve a entrar después de agregar el usuario al grupo `docker`.

Windows sirve para revisar el repositorio y preparar cambios, pero no puede cargar módulos Linux, usar `/proc` del host Linux, ejecutar cron de Linux ni adjuntar eBPF. En Windows no se deben tomar capturas de funcionamiento.

## 2. Infraestructura

El arranque recomendado se hace desde `Proyecto2`. El script levanta Valkey y Grafana, compila los componentes y verifica que el módulo y la sonda estén listos:

```bash
bash start.sh
```

Grafana queda disponible en `http://localhost:3001`. El usuario inicial es `admin` y la contraseña inicial es `admin`.

## 3. Imágenes

Construye las imágenes con los nombres que usa el cronjob:

```bash
docker build -t so1-pr2/go-client:latest docker-images/go-client
docker build -t so1-pr2/alpine-highcpu:latest docker-images/alpine-highcpu
docker build -t so1-pr2/alpine-lowconsumo:latest docker-images/alpine-lowconsumo
docker build -t so1-pr2/intruder:latest docker-images/intruder
```

El cronjob usa esas cuatro etiquetas y asigna la clase `low`, `high` o `intruder` mediante labels Docker.

## 4. Módulo de kernel

```bash
cd kernel-module
make
sudo insmod continfo.ko
head -n 10 /proc/continfo_pr2_so1_202303204
sudo rmmod continfo
cd ..
```

Para usar el script manual de carga del módulo:

```bash
chmod +x kernel-module/load_module.sh
sudo ./kernel-module/load_module.sh load
sudo ./kernel-module/load_module.sh unload
```

Si el módulo no carga, revisa el kernel y sus headers con `uname -r` y `ls /lib/modules/$(uname -r)/build`.

## 5. Parser de Go

```bash
cd daemon
go run . -proc /proc/continfo_pr2_so1_202303204 -once
go build -o daemon .
```

El modo `-once` facilita comprobar la lectura antes de dejar activo el ciclo periódico.

## 6. eBPF

```bash
cd ebpf-probe
make
sudo bpftool prog load kill_tracer.bpf.o /sys/fs/bpf/so1_pr2_kill_tracer autoattach
sudo bpftool link list
```

La sonda emite eventos con `bpf_printk` y el daemon los lee desde `trace_pipe`. No marques esta fase como validada hasta demostrar un evento real generado por una señal del daemon y ver la línea `SO1_PR2_KILL` en la traza del kernel.

## 7. Daemon

Después de ejecutar `start.sh`, compila y arranca el daemon como root para que pueda leer `/proc`, usar `sys_kill` y administrar el crontab root:

```bash
go build -C daemon -o daemon .
sudo ./daemon/daemon -interval 30s
```

El daemon instala el cronjob `*/2 * * * *`, carga de forma idempotente el módulo y la sonda, inicia el compose y elimina su entrada del crontab al recibir `SIGINT` o `SIGTERM`.

## 8. Verificación

```bash
cat /proc/continfo_pr2_so1_202303204
sudo crontab -l
sudo bpftool link list
```

Durante una eliminación deben aparecer en el daemon `SIGTERM enviado por el daemon` y `contenedor confirmado por eBPF`. Los contadores `containers:killed:count` y `ebpf:events:count` deben aumentar únicamente después de esa confirmación.

## 9. Cierre

```bash
Ctrl+C
bash stop.sh
```

El cronjob definitivo debe eliminarse antes de apagar el daemon. Durante las pruebas conviene revisar `crontab -l` y eliminar únicamente la entrada con el prefijo del Proyecto 2.
