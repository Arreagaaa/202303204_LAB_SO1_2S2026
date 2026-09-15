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

Desde `Proyecto2`:

```bash
docker compose -f infra/docker-compose.yml up -d
docker compose -f infra/docker-compose.yml ps
```

Grafana queda disponible en `http://localhost:3001`. El usuario inicial es `admin` y la contraseña inicial es `admin`; cámbiala al instalarlo en la VM.

## 3. Imágenes

Construye las imágenes con los nombres que usa el cronjob:

```bash
docker build -t so1-pr2/go-client:latest docker-images/go-client
docker build -t so1-pr2/alpine-highcpu:latest docker-images/alpine-highcpu
docker build -t so1-pr2/alpine-lowconsumo:latest docker-images/alpine-lowconsumo
docker build -t so1-pr2/intruder:latest docker-images/intruder
```

Antes de usar el cronjob, cambia sus imágenes para que correspondan a las etiquetas construidas. La versión inicial del script usa `alpine:3.20` para las dos variantes de Alpine y distingue el perfil con etiquetas Docker.

## 4. Módulo de kernel

```bash
cd kernel-module
make
sudo insmod continfo.ko
head -n 10 /proc/continfo_pr2_so1_202303204
sudo rmmod continfo
cd ..
```

Para usar el script:

```bash
chmod +x kernel-module/load_module.sh
./kernel-module/load_module.sh load
./kernel-module/load_module.sh unload
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
sudo bpftool prog load kill_tracer.bpf.o /sys/fs/bpf/so1_pr2_kill_tracer
```

La sonda emite eventos con `bpf_printk` y el daemon los lee desde `trace_pipe`. No marques esta fase como validada hasta demostrar un evento real generado por una señal del daemon y ver la línea `SO1_PR2_KILL` en la traza del kernel.

## 7. Cierre

```bash
sudo rmmod continfo || true
docker compose -f infra/docker-compose.yml down
```

El cronjob definitivo debe eliminarse antes de apagar el daemon. Durante las pruebas conviene revisar `crontab -l` y eliminar únicamente la entrada con el prefijo del Proyecto 2.
