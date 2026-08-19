# Proyecto 1 - Desarrollo, Conexion y Gestion de Contenedores en Entornos Virtualizados

**Curso:** Sistemas Operativos 1  
**Carnet:** 202303204  
**Fecha de inicio:** 30/07/2026  
**Fecha de calificacion:** 22/08/2026  

---

## 1. Descripcion del Proyecto

El proyecto consiste en el diseno e implementacion de un entorno virtualizado que integra el uso de maquinas virtuales (VMs) y contenedores, empleando tecnologias modernas como Docker, Containerd, Podman, Go y Zot. El objetivo es simular un entorno real de desarrollo distribuido, comprendiendo la contenerizacion, el almacenamiento de imagenes y la comunicacion entre servicios.

### 1.1 Objetivos

- Configurar 3 VMs con distintos propositos usando KVM
- Implementar APIs REST en Go con contenedores personalizados
- Configurar un registro privado de imagenes (Zot) para distribucion entre VMs
- Establecer comunicacion cruzada (REST/HTTP) entre servicios contenerizados
- Documentar toda la configuracion y el flujo del sistema

### 1.2 Tecnologias Utilizadas

| Componente | Tecnologia |
|---|---|
| Virtualizacion | KVM (virt-manager) |
| Lenguaje de APIs | Go (net/http) |
| Runtime VM1 | Containerd + nerdctl |
| Runtime VM2 | Podman |
| Runtime VM3 | Docker |
| Registro de imagenes | Zot |
| Formato de datos | JSON |
| Protocolo de comunicacion | REST/HTTP |

---

## 2. Arquitectura del Sistema

### 2.1 Distribucion de VMs

| Maquina Virtual | Runtime | Contenedores | IP |
|---|---|---|---|
| VM1 (vm1-containerd) | Containerd | API1, API2 | 192.0.2.11 |
| VM2 (vm2-podman) | Podman | API3 | 192.0.2.12 |
| VM3 (vm3-docker) | Docker | Zot (registro) | 192.0.2.13 |

### 2.2 Diagrama de Arquitectura

```
SISTEMA OPERATIVO LINUX (HOST)
         |
         | KVM
         v
+---------------------------+
| VM1 (192.0.2.11)      |
| containerd                |
|                           |
| API1 (:8081)  API2 (:8082)|
+---------------------------+
     |                 |
     | conexion        | conexion
     |                 v
     |        +---------------------------+
     |        | VM3 (192.0.2.13)      |
     |        | Docker                    |
     |        |                           |
     |        | Zot Registry (:5000)      |
     |        +---------------------------+
     |
     | conexion
     v
+---------------------------+
| VM2 (192.0.2.12)      |
| podman                    |
|                           |
| API3 (:8083)              |
+---------------------------+
```

### 2.3 Flujo de Imagenes

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

## 3. APIs Desarrolladas

### 3.1 Estructura de Endpoints

| API | Endpoint Principal | Endpoint de llamada 1 | Endpoint de llamada 2 |
|---|---|---|---|
| API1 (VM1, :8081) | GET /health | GET /api1/202303204/call-api2 | GET /api1/202303204/call-api3 |
| API2 (VM1, :8082) | GET /health | GET /api2/202303204/call-api1 | GET /api2/202303204/call-api3 |
| API3 (VM2, :8083) | GET /health | GET /api3/202303204/call-api1 | GET /api3/202303204/call-api2 |

### 3.2 Formato de Respuesta - /health

```json
{
  "status": "UP",
  "message": "API# is Ready",
  "timestamp": "2026-08-19T10:30:00Z",
  "VM": 1,
  "carnet": "202303204"
}
```

### 3.3 Formato de Respuesta - /call-api# (exito)

```json
{
  "apiname": "API#",
  "message": "The API# located on the VM# is working",
  "connection": true,
  "carnet": "202303204"
}
```

### 3.4 Formato de Respuesta - /call-api# (error)

```json
{
  "apiname": "API#",
  "message": "ERROR: The API# located on the VM# is not working",
  "connection": false,
  "carnet": "202303204"
}
```

### 3.5 Flujo de Comunicacion

Cuando un usuario solicita, por ejemplo, `GET /api2/202303204/call-api1`:

1. API2 recibe la peticion
2. API2 realiza una llamada HTTP interna a `GET /health` de API1
3. API2 analiza la respuesta:
   - Si `status == "UP"`: retorna JSON con `connection: true`
   - Si hay error o `status != "UP"`: retorna JSON con `connection: false`
4. El usuario recibe la respuesta formateada

---

## 4. Estructura del Repositorio

```
202303204_LAB_SO1_2S2026/
├── Proyecto1/
│   ├── docs/
│   │   └── proyecto-init.md
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

### 4.1 Nombres de Imagenes (obligatorio)

| Imagen | Nombre |
|---|---|
| API1 | API1-202303204 |
| API2 | API2-202303204 |
| API3 | API3-202303204 |

---

## 5. Estado Actual del Proyecto

### 5.1 Resumen de Avance

| Area | Estado |
|---|---|
| VM1 (containerd) | Completado - Runtime + nerdctl + buildkit listos |
| VM2 (podman) | Completado - Runtime listo |
| VM3 (Docker + Zot) | Completado - Runtime + registro Zot corriendo |
| Registro Zot accesible | Completado - Confirmado con push/pull cruzado |
| APIs en Go (API1, API2, API3) | Completado - Codigo desarrollado, compila correctamente |
| Dockerfiles | Completados - Multi-stage, golang:1.23-alpine |
| Build de imagenes | Pendiente |
| Push de imagenes al registro | Pendiente |
| Comunicacion cruzada entre APIs | Pendiente |
| Documentacion tecnica / manual | Pendiente |
| Repo GitHub privado + colaboradores | Pendiente |

### 5.2 Infraestructura Configurada

**VM1 - containerd:**
- Containerd v2.3.2 + nerdctl v2.3.4 + buildkitd
- Configuracion de registro inseguro en `/etc/containerd/certs.d/192.0.2.13:5000/hosts.toml`
- Prueba de pull desde Zot: OK

**VM2 - podman:**
- Podman v4.9.3 (rootless, cgroup v2, netavark)
- Configuracion de registro inseguro en `/etc/containers/registries.conf.d/insecure.conf`
- Prueba de pull desde Zot: OK

**VM3 - Docker + Zot:**
- Docker v29.7.2
- Zot v2.1.20 corriendo como contenedor Docker
- Puerto 5000 expuesto, config insecure-registries en daemon.json
- Prueba de push a Zot: OK

### 5.3 Proximos Pasos

1. Prender las 3 VMs y conectar por SSH
2. Build de imagenes con nerdctl (VM1) y podman (VM2)
3. Push de las 3 imagenes al registro Zot (VM3)
4. Pull + run de contenedores en VM1 y VM2
5. Probar comunicacion cruzada real entre APIs
6. Documentacion tecnica (manual + guia + diagrama + evidencia)
7. Configurar repo GitHub con colaboradores

---

## 6. Configuracion de Red

Todas las VMs y el host estan en la misma red NAT (192.0.2.x), lo que permite comunicacion directa entre ellas.

| Dispositivo | IP | Puerto |
|---|---|---|
| Host (Ubuntu) | 192.168.122.X | - |
| VM1 (containerd) | 192.0.2.11 | 8081 (API1), 8082 (API2) |
| VM2 (podman) | 192.0.2.12 | 8083 (API3) |
| VM3 (Docker + Zot) | 192.0.2.13 | 5000 (Zot) |

### 6.1 Variables de Entorno para APIs

Las APIs utilizan variables de entorno para configurar las URLs de destino:

| Variable | API1 | API2 | API3 |
|---|---|---|---|
| API1_URL | - | http://192.0.2.11:8081 | http://192.0.2.11:8081 |
| API2_URL | http://192.0.2.11:8082 | - | http://192.0.2.11:8082 |
| API3_URL | http://192.0.2.12:8083 | http://192.0.2.12:8083 | - |

---

## 7. Requisitos del Enunciado

### 7.1 Entregables

1. Codigo fuente en repositorio privado GitHub: `202303204_LAB_SO1_2S2026`
2. Manual tecnico y guia de instalacion
3. Capturas de prueba y evidencia funcional

### 7.2 Colaboradores del Repositorio

- `@JoseLorenzana272`
- `@KINGR0X`

### 7.3 Ponderacion del Proyecto

| Area | Puntos |
|---|---|
| Preparacion del Entorno | 20 |
| Desarrollo y Contenerizacion | 30 |
| Registros y Comunicacion Cruzada | 25 |
| Validacion y Documentacion | 15 |
| Habilidades Transversales | 10 |
| **Total** | **100** |

### 7.4 Cronograma

| Actividad | Fecha Inicio | Fecha Fin |
|---|---|---|
| Asignacion de Proyecto | 30/07/2026 | 20/08/2026 |
| Elaboracion | 30/07/2026 | 20/08/2026 |
| Calificacion | 22/08/2026 | 22/08/2026 |
