# Proyecto 1 - Comandos Ejecutados

Referencia de todos los comandos corridos durante la configuracion del entorno.

---

## 1. Host (Ubuntu)

### Configuracion de Git

```bash
git config --global user.email "tu-email@usac.com"
git config --global user.name "Javier"
```

### Build de imagenes Docker

```bash
# Navegar al directorio del proyecto
cd ~/Documentos/USAC/2S2026/SOPES1/LAB/202303204_LAB_SO1_2S2026/Proyecto1

# Build de las 3 APIs
docker build -t 192.0.2.13:5000/api1-202303204:latest ./api1
docker build -t 192.0.2.13:5000/api2-202303204:latest ./api2
docker build -t 192.0.2.13:5000/api3-202303204:latest ./api3
```

### Configuracion de daemon.json en host

```bash
# Configurar registro inseguro en el host
sudo nano /etc/docker/daemon.json
```

Contenido:
```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

```bash
# Reiniciar Docker
sudo systemctl restart docker
```

### Push de imagenes al registro Zot

```bash
docker push 192.0.2.13:5000/api1-202303204:latest
docker push 192.0.2.13:5000/api2-202303204:latest
docker push 192.0.2.13:5000/api3-202303204:latest
```

### Conexion SSH a las VMs

```bash
# VM1 (containerd)
ssh usuario@192.0.2.11

# VM2 (podman)
ssh usuario@192.0.2.12

# VM3 (Docker + Zot)
ssh usuario@192.0.2.13
```

---

## 2. VM1 - containerd (192.0.2.11)

### Instalacion de containerd y nerdctl

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar containerd
sudo apt install -y containerd

# Configurar containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml

# Reiniciar y habilitar containerd
sudo systemctl restart containerd
sudo systemctl enable containerd
sudo systemctl status containerd

# Instalar nerdctl (incluye buildkit y plugins CNI)
wget https://github.com/containerd/nerdctl/releases/download/v2.3.4/nerdctl-2.3.4-linux-amd64.tar.gz
sudo tar -xzf nerdctl-2.3.4-linux-amd64.tar.gz -C /usr/local/bin/

# Verificar nerdctl
nerdctl --version

# Probar nerdctl
sudo nerdctl run hello-world
```

### Configuracion de registro inseguro (Zot)

```bash
# Crear directorio de configuracion del registro
sudo mkdir -p /etc/containerd/certs.d/192.0.2.13:5000

# Crear archivo de configuracion
sudo nano /etc/containerd/certs.d/192.0.2.13:5000/hosts.toml
```

Contenido del archivo `hosts.toml`:

```toml
server = "http://192.0.2.13:5000"

[host."http://192.0.2.13:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
```

### Prueba de pull desde Zot

```bash
# Descargar imagen de prueba
sudo nerdctl --insecure-registry pull 192.0.2.13:5000/hello-world:test

# Verificar imagenes descargadas
sudo nerdctl images
```

### Pull y run de contenedores

```bash
# Descargar imagenes de APIs desde Zot
sudo nerdctl --insecure-registry pull 192.0.2.13:5000/api1-202303204:latest
sudo nerdctl --insecure-registry pull 192.0.2.13:5000/api2-202303204:latest

# Correr contenedores
sudo nerdctl --insecure-registry run -d --name api1 -p 8081:8081 192.0.2.13:5000/api1-202303204:latest
sudo nerdctl --insecure-registry run -d --name api2 -p 8082:8082 192.0.2.13:5000/api2-202303204:latest

# Verificar contenedores corriendo
sudo nerdctl ps

# Verificar logs
sudo nerdctl logs api1
sudo nerdctl logs api2
```

### Pruebas de endpoints

```bash
# Probar health de API1
curl http://192.0.2.11:8081/health
# Salida: {"status":"UP","message":"API1 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":1,"carnet":"202303204"}

# Probar health de API2
curl http://192.0.2.11:8082/health
# Salida: {"status":"UP","message":"API2 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":1,"carnet":"202303204"}

# Probar comunicacion API1 -> API2
curl http://192.0.2.11:8081/api1/202303204/call-api2
# Salida: {"apiname":"API2","message":"The API2 located on the VM1 is working","connection":true,"carnet":"202303204"}

# Probar comunicacion API1 -> API3
curl http://192.0.2.11:8081/api1/202303204/call-api3
# Salida: {"apiname":"API3","message":"The API3 located on the VM2 is working","connection":true,"carnet":"202303204"}

# Probar comunicacion API2 -> API1
curl http://192.0.2.11:8082/api2/202303204/call-api1
# Salida: {"apiname":"API1","message":"The API1 located on the VM1 is working","connection":true,"carnet":"202303204"}

# Probar comunicacion API2 -> API3
curl http://192.0.2.11:8082/api2/202303204/call-api3
# Salida: {"apiname":"API3","message":"The API3 located on the VM2 is working","connection":true,"carnet":"202303204"}
```

---

## 3. VM2 - podman (192.0.2.12)

### Instalacion de podman

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar podman
sudo apt install -y podman

# Verificar instalacion
podman --version
podman info
```

### Configuracion de registro inseguro (Zot)

```bash
# Crear archivo de configuracion
sudo nano /etc/containers/registries.conf.d/insecure.conf
```

Contenido del archivo `insecure.conf`:

```toml
[[registry]]
location = "192.0.2.13:5000"
insecure = true
```

### Prueba de pull desde Zot

```bash
# Descargar imagen de prueba
podman pull 192.0.2.13:5000/hello-world:test

# Verificar imagenes descargadas
podman images
```

### Pull y run de contenedor

```bash
# Descargar imagen de API3 desde Zot
podman pull 192.0.2.13:5000/api3-202303204:latest

# Correr contenedor
podman run -d --name api3 -p 8083:8083 192.0.2.13:5000/api3-202303204:latest

# Verificar contenedor corriendo
podman ps

# Verificar logs
podman logs api3
```

### Pruebas de endpoints

```bash
# Probar health de API3
curl http://192.0.2.12:8083/health
# Salida: {"status":"UP","message":"API3 is Ready","timestamp":"2026-08-20T01:09:19Z","VM":2,"carnet":"202303204"}

# Probar comunicacion API3 -> API1
curl http://192.0.2.12:8083/api3/202303204/call-api1
# Salida: {"apiname":"API1","message":"The API1 located on the VM1 is working","connection":true,"carnet":"202303204"}

# Probar comunicacion API3 -> API2
curl http://192.0.2.12:8083/api3/202303204/call-api2
# Salida: {"apiname":"API2","message":"The API2 located on the VM1 is working","connection":true,"carnet":"202303204"}
```

---

## 4. VM3 - Docker + Zot (192.0.2.13)

### Instalacion de Docker

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar Docker
sudo apt install -y docker.io

# Agregar usuario al grupo docker
sudo usermod -aG docker $USER

# Aplicar cambios de grupo (cerrar y abrir sesion, o ejecutar)
newgrp docker

# Verificar instalacion
docker --version
docker run hello-world
```

### Configuracion de Docker para registros inseguros

```bash
# Crear/editar daemon.json
sudo nano /etc/docker/daemon.json
```

Contenido del archivo `daemon.json`:

```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

```bash
# Reiniciar Docker para aplicar cambios
sudo systemctl restart docker
```

### Instalacion de Zot

```bash
# Crear directorios de configuracion y datos
mkdir -p ~/zot/config ~/zot/data

# Crear archivo de configuracion de Zot
nano ~/zot/config/config.json
```

Contenido del archivo `config.json`:

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

```bash
# Ejecutar Zot como contenedor Docker
docker run -d \
  --name zot-registry \
  --restart unless-stopped \
  -p 5000:5000 \
  -v ~/zot/config/config.json:/etc/zot/config.json \
  -v ~/zot/data:/var/lib/registry \
  ghcr.io/project-zot/zot-linux-amd64:latest

# Verificar que Zot esta corriendo
docker ps

# Verificar endpoint de catalogo
curl http://192.0.2.13:5000/v2/_catalog
# Salida esperada: {"repositories":[]}
```

### Prueba de push a Zot

```bash
# Tag de imagen de prueba
docker tag hello-world 192.0.2.13:5000/hello-world:test

# Push al registro
docker push 192.0.2.13:5000/hello-world:test

# Verificar catalogo
curl http://192.0.2.13:5000/v2/_catalog
# Salida esperada: {"repositories":["hello-world"]}
```

---

## 5. Pruebas de conectividad

### Verificar IPs de las VMs

```bash
# En cada VM, ejecutar:
ip addr show | grep inet
# o
hostname -I
```

IPs actuales:
- VM1: 192.0.2.11
- VM2: 192.0.2.12
- VM3: 192.0.2.13

### Probar conectividad entre VMs

```bash
# Desde VM1, probar conexion a VM2 y VM3
ping -c 3 192.0.2.12
ping -c 3 192.0.2.13

# Desde VM2, probar conexion a VM1 y VM3
ping -c 3 192.0.2.11
ping -c 3 192.0.2.13

# Desde VM3, probar conexion a VM1 y VM2
ping -c 3 192.0.2.11
ping -c 3 192.0.2.12
```

### Probar acceso a Zot desde cada VM

```bash
# Desde VM1
curl http://192.0.2.13:5000/v2/_catalog

# Desde VM2
curl http://192.0.2.13:5000/v2/_catalog

# Desde VM3 (local)
curl http://localhost:5000/v2/_catalog
```

---

## 6. Comandos utiles de referencia

### containerd / nerdctl

```bash
# Listar imagenes
sudo nerdctl images

# Listar contenedores
sudo nerdctl ps

# Ver logs de un contenedor
sudo nerdctl logs <nombre>

# Detener un contenedor
sudo nerdctl stop <nombre>

# Eliminar un contenedor
sudo nerdctl rm <nombre>

# Eliminar una imagen
sudo nerdctl rmi <imagen>
```

### podman

```bash
# Listar imagenes
podman images

# Listar contenedores
podman ps

# Ver logs de un contenedor
podman logs <nombre>

# Detener un contenedor
podman stop <nombre>

# Eliminar un contenedor
podman rm <nombre>

# Eliminar una imagen
podman rmi <imagen>
```

### Docker

```bash
# Listar imagenes
docker images

# Listar contenedores
docker ps

# Ver logs de un contenedor
docker logs <nombre>

# Detener un contenedor
docker stop <nombre>

# Eliminar un contenedor
docker rm <nombre>

# Eliminar una imagen
docker rmi <imagen>
```

---

## 7. Estructura de archivos de configuracion

### VM1 - /etc/containerd/certs.d/192.0.2.13:5000/hosts.toml

```toml
server = "http://192.0.2.13:5000"

[host."http://192.0.2.13:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
```

### VM2 - /etc/containers/registries.conf.d/insecure.conf

```toml
[[registry]]
location = "192.0.2.13:5000"
insecure = true
```

### VM3 - /etc/docker/daemon.json

```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

### VM3 - ~/zot/config/config.json

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
