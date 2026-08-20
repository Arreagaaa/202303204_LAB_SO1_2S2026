# Guia de Instalacion - Proyecto 1: Desarrollo, Conexion y Gestion de Contenedores en Entornos Virtualizados

**Curso:** Sistemas Operativos 1  
**Carnet:** 202303204  
**Fecha:** 20/08/2026  

---

## 1. Requisitos Previos

### 1.1 Host (Ubuntu)

- Sistema operativo Linux con virtualizacion activada (KVM)
- Docker instalado y configurado
- Conectividad a la red NAT (192.0.2.x)
- Acceso SSH a las 3 VMs

### 1.2 Maquinas Virtuales

| VM | SO | RAM | CPUs | Disco |
|---|---|---|---|---|
| VM1 | Ubuntu Server 24.04.4 LTS | 2048 MiB | 2 | 15 GiB |
| VM2 | Ubuntu Server 24.04.4 LTS | 2048 MiB | 2 | 15 GiB |
| VM3 | Ubuntu Server 24.04.4 LTS | 2048 MiB | 2 | 15 GiB |

---

## 2. Configuracion del Host

### 2.1 Configuracion de Git

```bash
git config --global user.email "tu-email@usac.com"
git config --global user.name "Javier"
```

### 2.2 Configuracion de Docker para Registros Inseguros

Zot funciona sin HTTPS, por lo que Docker necesita saber que puede conectarse a ese registro sin encriptacion.

```bash
sudo nano /etc/docker/daemon.json
```

Agregar el siguiente contenido:

```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

Reiniciar Docker para aplicar los cambios:

```bash
sudo systemctl restart docker
```

### 2.3 Build de Imagenes

Navegar al directorio del proyecto y construir las 3 imagenes:

```bash
cd ~/Documentos/USAC/2S2026/SOPES1/LAB/202303204_LAB_SO1_2S2026/Proyecto1

docker build -t 192.0.2.13:5000/API1-202303204:latest ./api1
docker build -t 192.0.2.13:5000/API2-202303204:latest ./api2
docker build -t 192.0.2.13:5000/API3-202303204:latest ./api3
```

### 2.4 Push de Imagenes al Registro Zot

Subir las imagenes construidas al registro de VM3:

```bash
docker push 192.0.2.13:5000/API1-202303204:latest
docker push 192.0.2.13:5000/API2-202303204:latest
docker push 192.0.2.13:5000/API3-202303204:latest
```

### 2.5 Verificacion del Catalogo

```bash
curl http://192.0.2.13:5000/v2/_catalog
# Salida esperada: {"repositories":["api1-202303204","api2-202303204","api3-202303204"]}
```

![Catalogo de Zot](images/06-catalogo-zot.png)

---

## 3. Configuracion de VM1 (Containerd)

### 3.1 Conexion

```bash
ssh usuario@192.0.2.11
```

Contrasena: `1234javier`

### 3.2 Instalacion de Containerd

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y containerd

# Configurar containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml

# Reiniciar y habilitar
sudo systemctl restart containerd
sudo systemctl enable containerd
```

### 3.3 Instalacion de Nerdctl

Nerdctl es un cliente de contenedores para containerd con interfaz similar a Docker.

```bash
wget https://github.com/containerd/nerdctl/releases/download/v2.3.4/nerdctl-2.3.4-linux-amd64.tar.gz
sudo tar -xzf nerdctl-2.3.4-linux-amd64.tar.gz -C /usr/local/bin/
nerdctl --version
```

### 3.4 Configuracion de Registro Inseguro

Para que nerdctl pueda comunicarse con Zot sin HTTPS:

```bash
sudo mkdir -p /etc/containerd/certs.d/192.0.2.13:5000
sudo nano /etc/containerd/certs.d/192.0.2.13:5000/hosts.toml
```

Agregar el siguiente contenido:

```toml
server = "http://192.0.2.13:5000"

[host."http://192.0.2.13:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
```

### 3.5 Pull de Imagenes

Descargar las imagenes de API1 y API2 desde Zot:

```bash
sudo nerdctl --insecure-registry pull 192.0.2.13:5000/API1-202303204:latest
sudo nerdctl --insecure-registry pull 192.0.2.13:5000/API2-202303204:latest
```

### 3.6 Ejecucion de Contenedores

```bash
sudo nerdctl --insecure-registry run -d --name api1 -p 8081:8081 192.0.2.13:5000/API1-202303204:latest
sudo nerdctl --insecure-registry run -d --name api2 -p 8082:8082 192.0.2.13:5000/API2-202303204:latest
```

### 3.7 Verificacion

```bash
sudo nerdctl ps
```

Salida esperada:

```
CONTAINER ID    IMAGE                                         COMMAND     STATUS    PORTS                     NAMES
1ff1507c8181    192.0.2.13:5000/API2-202303204:latest    "./main"    Up        0.0.0.0:8082->8082/tcp    api2
d705c32bb8f2    192.0.2.13:5000/API1-202303204:latest    "./main"    Up        0.0.0.0:8081->8081/tcp    api1
```

![nerdctl ps en VM1](images/02-vm1-nerdctl-ps.png)

---

## 4. Configuracion de VM2 (Podman)

### 4.1 Conexion

```bash
ssh usuario@192.0.2.12
```

Contrasena: `1234javier`

### 4.2 Instalacion de Podman

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y podman

podman --version
podman info
```

### 4.3 Configuracion de Registro Inseguro

Para que podman pueda comunicarse con Zot sin HTTPS:

```bash
sudo nano /etc/containers/registries.conf.d/insecure.conf
```

Agregar el siguiente contenido:

```toml
[[registry]]
location = "192.0.2.13:5000"
insecure = true
```

### 4.4 Pull de Imagenes

Descargar la imagen de API3 desde Zot:

```bash
podman pull 192.0.2.13:5000/API3-202303204:latest
```

### 4.5 Ejecucion de Contenedor

```bash
podman run -d --name api3 -p 8083:8083 192.0.2.13:5000/API3-202303204:latest
```

### 4.6 Verificacion

```bash
podman ps
```

Salida esperada:

```
CONTAINER ID  IMAGE                                       COMMAND     CREATED         STATUS         PORTS                   NAMES
d1562eec1577  192.0.2.13:5000/API3-202303204:latest  ./main      Up 10 seconds  0.0.0.0:8083->8083/tcp  api3
```

![podman ps en VM2](images/03-vm2-podman-ps.png)

---

## 5. Configuracion de VM3 (Docker + Zot)

### 5.1 Conexion

```bash
ssh usuario@192.0.2.13
```

Contrasena: `1234javier`

### 5.2 Instalacion de Docker

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io

sudo usermod -aG docker $USER
newgrp docker

docker --version
docker run hello-world
```

### 5.3 Configuracion de Docker para Registros Inseguros

```bash
sudo nano /etc/docker/daemon.json
```

Agregar el siguiente contenido:

```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

```bash
sudo systemctl restart docker
```

### 5.4 Instalacion de Zot

Zot es un registro de imagenes OCI que se ejecuta como contenedor Docker.

```bash
# Crear directorios de configuracion y datos
mkdir -p ~/zot/config ~/zot/data

# Crear archivo de configuracion
nano ~/zot/config/config.json
```

Agregar el siguiente contenido:

```json
{
  "storage": {
    "rootDirectory": "/var/lib/registry"
  },
  "http": {
    "address": "0.0.0.0",
    "port": "5000"
  },
  "log": {
    "level": "info"
  }
}
```

Ejecutar Zot:

```bash
docker run -d \
  --name zot-registry \
  --restart unless-stopped \
  -p 5000:5000 \
  -v ~/zot/config/config.json:/etc/zot/config.json \
  -v ~/zot/data:/var/lib/registry \
  ghcr.io/project-zot/zot-linux-amd64:latest
```

### 5.5 Verificacion

```bash
# Verificar que Zot esta corriendo
docker ps

# Verificar endpoint de catalogo
curl http://192.0.2.13:5000/v2/_catalog
# Salida esperada: {"repositories":[]}
```

![docker ps en VM3](images/04-vm3-docker-ps-zot.png)

---

## 6. Pruebas de Comunicacion

### 6.1 Prueba de Health Endpoints

Desde el host o cualquier VM:

```bash
# API1
curl http://192.0.2.11:8081/health
# Salida: {"status":"UP","message":"API1 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":1,"carnet":"202303204"}

# API2
curl http://192.0.2.11:8082/health
# Salida: {"status":"UP","message":"API2 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":1,"carnet":"202303204"}

# API3
curl http://192.0.2.12:8083/health
# Salida: {"status":"UP","message":"API3 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":2,"carnet":"202303204"}
```

### 6.2 Prueba de Comunicacion Cruzada

```bash
# API1 -> API2 (misma VM)
curl http://192.0.2.11:8081/api1/202303204/call-api2
# Salida: {"apiname":"API2","message":"The API2 located on the VM1 is working","connection":true,"carnet":"202303204"}

# API1 -> API3 (VM2)
curl http://192.0.2.11:8081/api1/202303204/call-api3
# Salida: {"apiname":"API3","message":"The API3 located on the VM2 is working","connection":true,"carnet":"202303204"}

# API2 -> API1 (misma VM)
curl http://192.0.2.11:8082/api2/202303204/call-api1
# Salida: {"apiname":"API1","message":"The API1 located on the VM1 is working","connection":true,"carnet":"202303204"}

# API2 -> API3 (VM2)
curl http://192.0.2.11:8082/api2/202303204/call-api3
# Salida: {"apiname":"API3","message":"The API3 located on the VM2 is working","connection":true,"carnet":"202303204"}

# API3 -> API1 (VM1)
curl http://192.0.2.12:8083/api3/202303204/call-api1
# Salida: {"apiname":"API1","message":"The API1 located on the VM1 is working","connection":true,"carnet":"202303204"}

# API3 -> API2 (VM1)
curl http://192.0.2.12:8083/api3/202303204/call-api2
# Salida: {"apiname":"API2","message":"The API2 located on the VM1 is working","connection":true,"carnet":"202303204"}
```

![Pruebas de comunicacion entre APIs](images/05-pruebas-comunicacion.png)

### 6.3 Prueba de Error (API Apagada)

Para probar el manejo de errores, detener un contenedor y verificar que la respuesta sea `connection: false`:

```bash
# En VM1, detener API1
sudo nerdctl stop api1

# Desde VM2 o el host, probar comunicacion con API1
curl http://192.0.2.11:8081/api2/202303204/call-api1
# Salida esperada: {"apiname":"API1","message":"ERROR: The API1 located on the VM1 is not working","connection":false,"carnet":"202303204"}

# Volver a iniciar API1
sudo nerdctl start api1
```

---

## 7. Comandos Utiles de Referencia

### 7.1 Containerd / Nerdctl

```bash
sudo nerdctl images                    # Listar imagenes
sudo nerdctl ps                        # Listar contenedores
sudo nerdctl logs <nombre>             # Ver logs
sudo nerdctl stop <nombre>             # Detener contenedor
sudo nerdctl rm <nombre>               # Eliminar contenedor
sudo nerdctl rmi <imagen>              # Eliminar imagen
```

### 7.2 Podman

```bash
podman images                          # Listar imagenes
podman ps                              # Listar contenedores
podman logs <nombre>                   # Ver logs
podman stop <nombre>                   # Detener contenedor
podman rm <nombre>                     # Eliminar contenedor
podman rmi <imagen>                    # Eliminar imagen
```

### 7.3 Docker

```bash
docker images                          # Listar imagenes
docker ps                              # Listar contenedores
docker logs <nombre>                   # Ver logs
docker stop <nombre>                   # Detener contenedor
docker rm <nombre>                     # Eliminar contenedor
docker rmi <imagen>                    # Eliminar imagen
```

---

## 8. Solucion de Problemas

### 8.1 Error de HTTPS en Push/Pull

**Problema:** `http: server gave HTTP response to HTTPS client`

**Solucion:** Verificar que el archivo `daemon.json` (host/VM3) o `hosts.toml` (VM1) o `insecure.conf` (VM2) este configurado correctamente con la IP del registro.

### 8.2 Contenedor No Inicia

**Problema:** El contenedor se crea pero no permanece corriendo.

**Solucion:** Verificar los logs con `sudo nerdctl logs <nombre>`, `podman logs <nombre>` o `docker logs <nombre>`.

### 8.3 No hay Comunicacion entre APIs

**Problema:** Los curls de call-api# retornan error.

**Solucion:** Verificar que todos los contenedores esten corriendo y que la red NAT este funcionando (probar con `ping`).

### 8.4 Puerto en Uso

**Problema:** `Bind for 0.0.0.0:8081 failed: port is already allocated`

**Solucion:** Detener el contenedor que esta usando ese puerto o eliminar contenedores anteriores con el mismo nombre.
