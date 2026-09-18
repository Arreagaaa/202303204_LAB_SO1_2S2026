# Guía rápida para la presentación

Esta guía corresponde al Proyecto 2 de telemetría de contenedores, carnet 202303204.

## Antes de salir

Verificar que el último commit incluya el código, la documentación y las imágenes de `docs/images/`.

## Quince minutos antes

Abrir la primera terminal y ejecutar:

```bash
cd /home/javier/Documentos/USAC/2S2026/SOPES1/LAB/202303204_LAB_SO1_2S2026/Proyecto2
bash stop.sh
bash start.sh
go build -C daemon -o daemon .
sudo ./daemon/daemon -interval 30s
```

No cerrar esta terminal. Aquí se mostrarán los mensajes del daemon y las confirmaciones eBPF.

Esperar unos minutos para que el cronjob cree contenedores y el daemon procese al menos una lectura.

## Segunda terminal

Abrir otra terminal en el mismo directorio:

```bash
cd /home/javier/Documentos/USAC/2S2026/SOPES1/LAB/202303204_LAB_SO1_2S2026/Proyecto2
```

## Vuelta 1: módulo de kernel

```bash
lsmod | grep continfo
cat /proc/continfo_pr2_so1_202303204
```

Mostrar que aparecen:

- memoria total;
- memoria libre;
- memoria usada;
- PID;
- nombre;
- comando o identificador;
- VSZ;
- RSS;
- porcentaje de memoria;
- porcentaje de CPU.

## Vuelta 2: procesos y cronjob

Mostrar los contenedores activos:

```bash
docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Status}}"
```

Relacionar el nombre Docker, el PID y la fila correspondiente de `/proc`:

```bash
docker ps --format '{{.ID}} {{.Names}}' | while read -r id name; do
  pid=$(docker inspect -f '{{.State.Pid}}' "$id")
  printf '%s pid=%s: ' "$name" "$pid"
  grep "^${pid}," /proc/continfo_pr2_so1_202303204 || true
done
```

El PID de Docker debe corresponder con una fila del archivo `/proc`. El módulo
reporta `task->comm` en `NAME` y `CMD_OR_CONTAINER_ID`; el daemon agrega el
nombre, ID y clase Docker al cruzar ese PID con `docker inspect`.

El daemon se ejecuta con `sudo`, por lo que el cronjob queda instalado para root:

```bash
crontab -l
sudo crontab -l
```

La segunda salida debe contener una línea similar a:

```text
*/2 * * * * /bin/bash .../Proyecto2/cronjob/spawn_containers.sh # proyecto2-so1
```

El cronjob genera tres contenedores `low`, dos `high` y un `intruder` por lote. El daemon conserva los mínimos y no elimina Grafana.

## Vuelta 3: daemon y Valkey

Verificar que el daemon esté vivo:

```bash
ps aux | grep '[d]aemon/daemon'
```

Consultar Valkey desde el contenedor:

```bash
docker exec so1-pr2-valkey valkey-cli KEYS '*'
docker exec so1-pr2-valkey valkey-cli GET system:ram:current
docker exec so1-pr2-valkey valkey-cli GET system:ram:used
docker exec so1-pr2-valkey valkey-cli GET containers:killed:count
docker exec so1-pr2-valkey valkey-cli GET ebpf:events:count
```

Para mostrar los rankings:

```bash
docker exec so1-pr2-valkey valkey-cli HGETALL containers:top:ram
docker exec so1-pr2-valkey valkey-cli HGETALL containers:top:cpu
```

## Vuelta 4: eBPF

Mostrar que la sonda está adjunta:

```bash
sudo bpftool link list
```

Debe aparecer:

```text
tracepoint sys_enter_kill
```

En la primera terminal esperar una eliminación. La secuencia esperada es:

```text
SIGTERM enviado por el daemon
contenedor confirmado por eBPF
contenedor detenido
contenedor eliminado
```

Los mensajes tienen timestamp reciente. Después mostrar la confirmación en Valkey:

```bash
docker exec so1-pr2-valkey valkey-cli GET containers:killed:count
docker exec so1-pr2-valkey valkey-cli GET ebpf:events:count
docker exec so1-pr2-valkey valkey-cli LRANGE containers:killed:history -3 -1
```

Los dos contadores deben coincidir. El registro solo se escribe después de la confirmación eBPF.

## Vuelta 5: Grafana

Abrir:

```text
http://localhost:3001
```

Mostrar el dashboard `SO1 Proyecto 2 - Telemetria` con:

- RAM total;
- RAM usada;
- memoria libre;
- contenedores eliminados;
- eventos eBPF;
- Top 5 por RAM;
- Top 5 por CPU.

## Preguntas rápidas

### ¿Por qué se usa `/proc`?

Porque el módulo de kernel expone allí la información de memoria y procesos mediante `seq_file`.

### ¿Por qué el módulo muestra `unknown` en `CLASS`?

El módulo conoce los procesos del kernel, pero las clases `low`, `high` e `intruder` provienen de las etiquetas Docker. El daemon cruza los PID con Docker y asigna la clase correcta.

### ¿Cómo se protegen los mínimos?

La política conserva tres contenedores `low` y dos `high`. Grafana se excluye explícitamente de los candidatos.

### ¿Qué confirma eBPF?

Confirma que el daemon envió una señal `SIGTERM` mediante `sys_kill` al PID del contenedor.

### ¿Cuándo se registra un contenedor eliminado?

Después de que el daemon recibe la confirmación eBPF. Entonces lo detiene, lo elimina y actualiza Valkey.

### ¿Qué significan valores de CPU cercanos a `10000`?

El daemon conserva dos decimales multiplicados por 100. Por ejemplo, `9980` representa aproximadamente `99.80%`.

## Cierre después de la presentación

En la primera terminal presionar `Ctrl+C` y luego ejecutar:

```bash
bash stop.sh
```

Esto elimina el cronjob, los contenedores de prueba, el módulo, la sonda, Grafana y Valkey.

La actividad de Rust es opcional y no forma parte de la calificación obligatoria del proyecto.
