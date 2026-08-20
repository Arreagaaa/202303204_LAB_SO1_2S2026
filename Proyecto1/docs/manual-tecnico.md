# Manual Tecnico - Proyecto 1: Desarrollo, Conexion y Gestion de Contenedores en Entornos Virtualizados

**Curso:** Sistemas Operativos 1  
**Carnet:** 202303204  
**Fecha:** 20/08/2026  

---

## 1. Introduccion

Este documento describe la arquitectura tecnica del proyecto, incluyendo la configuracion de las maquinas virtuales, los contenedores implementados, la comunicacion entre servicios y el registro de imagenes utilizado.

El proyecto implementa un entorno distribuido con 3 maquinas virtuales, cada una ejecutando un runtime de contenedores diferente (Containerd, Podman, Docker), comunicandose entre si mediante APIs REST desarrolladas en Go.

---

## 2. Arquitectura del Sistema

### 2.1 Diagrama de Arquitectura

![Arquitectura del Sistema](images/01-arquitectura-sistema.png)

```
SISTEMA OPERATIVO LINUX (HOST)
         |
         | KVM (Virtualizacion)
         v
+---------------------------+
| VM1 (192.0.2.11)      |
| Runtime: Containerd       |
|                           |
| API1 (:8081)  API2 (:8082)|
+---------------------------+
     |                 |
     | conexion        | conexion
     |                 v
     |        +---------------------------+
     |        | VM3 (192.0.2.13)      |
     |        | Runtime: Docker            |
     |        |                            |
     |        | Zot Registry (:5000)       |
     |        | Registro de imagenes       |
     |        +---------------------------+
     |
     | conexion
     v
+---------------------------+
| VM2 (192.0.2.12)      |
| Runtime: Podman           |
|                           |
| API3 (:8083)              |
+---------------------------+
```

### 2.2 Flujo de Imagenes

```
HOST (Ubuntu)               VM3 (Zot)              VM1 (containerd)    VM2 (podman)
     |                          |                        |                  |
 docker build               Zot corriendo             containerd         podman
 docker push ------------> push a :5000               |                  |
                           Zot almacena                |                  |
                                 |                pull desde Zot      pull desde Zot
                                 |                 nerdctl run        podman run
```

---

## 3. Configuracion de Maquinas Virtuales

### 3.1 VM1 - containerd

| Campo | Valor |
|---|---|
| Nombre (virt-manager) | `vm1-containerd` |
| Sistema Operativo | Ubuntu Server 24.04.4 LTS |
| RAM | 2048 MiB |
| CPUs | 2 |
| Disco | 15 GiB |
| Red | NAT (192.0.2.x) |
| IP | `192.0.2.11` |
| Runtime | Containerd v2.3.2 + nerdctl v2.3.4 |
| Contenedores | API1, API2 |

**Componentes instalados:**
- Containerd (via apt)
- nerdctl v2.3.4 (paquete nerdctl-full)
- buildkitd (servicio systemd activo)
- Plugins CNI

**Configuracion de registro:**
- Archivo: `/etc/containerd/certs.d/192.0.2.13:5000/hosts.toml`

```toml
server = "http://192.0.2.13:5000"

[host."http://192.0.2.13:5000"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
```

**Contenedores corriendo:**

![nerdctl ps en VM1](images/02-vm1-nerdctl-ps.png)

### 3.2 VM2 - podman

| Campo | Valor |
|---|---|
| Nombre (virt-manager) | `vm2-podman` |
| Sistema Operativo | Ubuntu Server 24.04.4 LTS |
| RAM | 2048 MiB |
| CPUs | 2 |
| Disco | 15 GiB |
| Red | NAT (192.0.2.x) |
| IP | `192.0.2.12` |
| Runtime | Podman v4.9.3 |
| Contenedores | API3 |

**Componentes instalados:**
- Podman v4.9.3 (rootless, cgroup v2, netavark)

**Configuracion de registro:**
- Archivo: `/etc/containers/registries.conf.d/insecure.conf`

```toml
[[registry]]
location = "192.0.2.13:5000"
insecure = true
```

**Contenedor corriendo:**

![podman ps en VM2](images/03-vm2-podman-ps.png)

### 3.3 VM3 - Docker + Zot

| Campo | Valor |
|---|---|
| Nombre (virt-manager) | `vm3-docker` |
| Sistema Operativo | Ubuntu Server 24.04.4 LTS |
| RAM | 2048 MiB |
| CPUs | 2 |
| Disco | 15 GiB |
| Red | NAT (192.0.2.x) |
| IP | `192.0.2.13` |
| Runtime | Docker v29.7.2 |
| Contenedores | Zot (registro de imagenes) |

**Componentes instalados:**
- Docker v29.7.2
- Zot v2.1.20 (contenedor Docker)

**Configuracion Docker:**
- Archivo: `/etc/docker/daemon.json`

```json
{
  "insecure-registries": ["192.0.2.13:5000"]
}
```

**Configuracion Zot:**
- Config: `~/zot/config/config.json` montado en `/etc/zot/config.json`
- Storage: `~/zot/data` montado en `/var/lib/registry`
- Puerto: `5000` (0.0.0.0:5000->5000/tcp)
- Reinicio: `--restart unless-stopped`

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

**Zot corriendo:**

![docker ps en VM3](images/04-vm3-docker-ps-zot.png)

---

## 4. APIs Desarrolladas

### 4.1 Especificacion de Endpoints

| API | Ubicacion | Endpoint Principal | Endpoint de llamada 1 | Endpoint de llamada 2 |
|---|---|---|---|---|
| API1 | VM1 (:8081) | GET /health | GET /api1/202303204/call-api2 | GET /api1/202303204/call-api3 |
| API2 | VM1 (:8082) | GET /health | GET /api2/202303204/call-api1 | GET /api2/202303204/call-api3 |
| API3 | VM2 (:8083) | GET /health | GET /api3/202303204/call-api1 | GET /api3/202303204/call-api2 |

### 4.2 Formato de Respuesta - /health

```json
{
  "status": "UP",
  "message": "API# is Ready",
  "timestamp": "2026-08-20T01:09:19Z",
  "VM": 1,
  "carnet": "202303204"
}
```

### 4.3 Formato de Respuesta - /call-api# (exito)

```json
{
  "apiname": "API#",
  "message": "The API# located on the VM# is working",
  "connection": true,
  "carnet": "202303204"
}
```

### 4.4 Formato de Respuesta - /call-api# (error)

```json
{
  "apiname": "API#",
  "message": "ERROR: The API# located on the VM# is not working",
  "connection": false,
  "carnet": "202303204"
}
```

### 4.5 Flujo de Comunicacion

Cuando un usuario solicita `GET /api2/202303204/call-api1`:

1. API2 recibe la peticion HTTP
2. API2 realiza una llamada interna a `GET /health` de API1
3. API2 analiza la respuesta de API1:
   - Si `status == "UP"`: retorna JSON con `connection: true`
   - Si hay error o `status != "UP"`: retorna JSON con `connection: false`
4. El usuario recibe la respuesta formateada

**Pruebas de comunicacion:**

![Pruebas de comunicacion entre APIs](images/05-pruebas-comunicacion.png)

### 4.6 Variables de Entorno

Las APIs utilizan variables de entorno para configurar las URLs de destino:

| Variable | API1 | API2 | API3 |
|---|---|---|---|
| API1_URL | - | http://192.0.2.11:8081 | http://192.0.2.11:8081 |
| API2_URL | http://192.0.2.11:8082 | - | http://192.0.2.11:8082 |
| API3_URL | http://192.0.2.12:8083 | http://192.0.2.12:8083 | - |

---

## 5. Registro de Imagenes (Zot)

### 5.1 Descripcion

Zot es un registro de imagenes de contenedores OCI que se ejecuta como contenedor Docker en VM3. Almacena las imagenes de las 3 APIs y permite su distribucion entre las VMs.

### 5.2 Imagenes Almacenadas

| Imagen | Tag | Ubicacion |
|---|---|---|
| API1-202303204 | latest | 192.0.2.13:5000 |
| API2-202303204 | latest | 192.0.2.13:5000 |
| API3-202303204 | latest | 192.0.2.13:5000 |

### 5.3 Verificacion del Catalogo

```bash
curl http://192.0.2.13:5000/v2/_catalog
# Salida: {"repositories":["api1-202303204","api2-202303204","api3-202303204"]}
```

![Catalogo de Zot](images/06-catalogo-zot.png)

---

## 6. Configuracion de Red

Todas las VMs y el host estan en la misma red NAT (192.0.2.x), lo que permite comunicacion directa entre ellas.

| Dispositivo | IP | Puertos |
|---|---|---|
| Host (Ubuntu) | 192.168.122.X | - |
| VM1 (containerd) | 192.0.2.11 | 8081 (API1), 8082 (API2) |
| VM2 (podman) | 192.0.2.12 | 8083 (API3) |
| VM3 (Docker + Zot) | 192.0.2.13 | 5000 (Zot) |

### 6.1 Pruebas de Conectividad

```bash
# Verificar conexion desde host a VM1
ping -c 3 192.0.2.11

# Verificar conexion desde host a VM2
ping -c 3 192.0.2.12

# Verificar conexion desde host a VM3
ping -c 3 192.0.2.13
```

---

## 7. Estructura del Repositorio

```
202303204_LAB_SO1_2S2026/
├── Proyecto1/
│   ├── docs/
│   │   ├── manual-tecnico.md
│   │   ├── guia-instalacion.md
│   │   └── images/
│   │       ├── 01-arquitectura-sistema.png
│   │       ├── 02-vm1-nerdctl-ps.png
│   │       ├── 03-vm2-podman-ps.png
│   │       ├── 04-vm3-docker-ps-zot.png
│   │       ├── 05-pruebas-comunicacion.png
│   │       └── 06-catalogo-zot.png
│   ├── api1/
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── api2/
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── api3/
│       ├── main.go
│       ├── go.mod
│       └── Dockerfile
└── README.md
```

---

## 8. Runtimes de Contenedores

### 8.1 Comparativa

| Caracteristica | Containerd | Podman | Docker |
|---|---|---|---|
| Daemon | Si (containerd) | No (daemonless) | Si (dockerd) |
| CLI | nerdctl | podman | docker |
| OCI compliant | Si | Si | Si |
| Rootless | Si | Si (por defecto) | Si (configurable) |
| Uso en este proyecto | VM1 (API1, API2) | VM2 (API3) | VM3 (Zot) |

### 8.2 Containerd (VM1)

Containerd es un runtime de contenedores de bajo nivel que maneja el ciclo de vida de los contenedores. Se utiliza nerdctl como cliente CLI para interactuar con containerd de forma similar a Docker.

### 8.3 Podman (VM2)

Podman es un runtime de contenedores daemonless, lo que significa que no requiere un proceso central corriendo en segundo plano. Cada contenedor se ejecuta de forma independiente.

### 8.4 Docker (VM3)

Docker es el runtime de contenedores mas utilizado en la industria. En este proyecto se usa exclusivamente para ejecutar el registro de imagenes Zot.

---

## 9. Conclusiones Tecnicas

1. **Infraestructura:** Se configuro exitosamente un entorno con 3 VMs usando KVM, cada una con un runtime de contenedores diferente.

2. **Desarrollo de APIs:** Las 3 APIs fueron desarrolladas en Go utilizando la libreria estandar `net/http`, implementando los endpoints requeridos por el enunciado.

3. **Registro de imagenes:** Zot funciona como registro privado, permitiendo la distribucion de imagenes entre las VMs mediante push/pull.

4. **Comunicacion:** Las APIs se comunican entre si mediante HTTP, verificando el estado de las demas APIs a traves del endpoint `/health`.

5. **Portabilidad:** Las imagenes construidas son OCI-compliant, lo que permite ejecutarlas en cualquier runtime compatible (containerd, podman, docker).

---

## 10. Referencias

- [Containerd](https://containerd.io/)
- [Nerdctl](https://github.com/containerd/nerdctl)
- [Podman](https://podman.io/)
- [Docker](https://www.docker.com/)
- [Zot](https://github.com/project-zot/zot)
- [Go net/http](https://pkg.go.dev/net/http)
