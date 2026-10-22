"""Comparación local y aislada de Valkey primario solo frente a primario + réplica.

Ejecutar: python3 scripts/compare-valkey-local.py
Requiere Docker local e imagen valkey/valkey:8. No toca Kubernetes ni GCP.
"""

import json
import socket
import statistics
import subprocess
import time

IMAGE = "valkey/valkey:8"
PRIMARY = "mumnk8s-valkey-bench-primary"
REPLICA = "mumnk8s-valkey-bench-replica"
NETWORK = "mumnk8s-valkey-bench"
PORT = 16379
REPLICA_PORT = 16380
SAMPLES = 1000


def docker(*args):
    return subprocess.run(["docker", *args], check=True, capture_output=True, text=True).stdout.strip()


def resp(port, *parts):
    payload = f"*{len(parts)}\r\n".encode() + b"".join(
        f"${len(str(part).encode())}\r\n{part}\r\n".encode() for part in parts
    )
    with socket.create_connection(("127.0.0.1", port), timeout=3) as conn:
        conn.settimeout(3)
        conn.sendall(payload)
        with conn.makefile("rb") as stream:
            prefix = stream.read(1)
            line = stream.readline().strip()
            if prefix == b"-":
                raise RuntimeError(line.decode())
            if prefix in (b"+", b":"):
                return line.decode()
            if prefix == b"$":
                return stream.read(int(line) + 2)[:-2].decode()
            raise RuntimeError(f"Respuesta RESP inesperada: {prefix!r}")


def ready(port):
    for _ in range(50):
        try:
            if resp(port, "PING") == "PONG":
                return
        except (OSError, RuntimeError):
            time.sleep(0.1)
    raise RuntimeError(f"Valkey en puerto {port} no inició")


def trial(name, replicas):
    key = "reports:CHN"
    resp(PORT, "DEL", key)
    times = []
    for i in range(SAMPLES):
        report = json.dumps({"country": "CHN", "warplanes_in_air": i % 51,
                             "warships_in_water": i % 31, "timestamp": "2026-10-22T00:00:00Z"})
        start = time.perf_counter()
        resp(PORT, "RPUSH", key, report)
        times.append((time.perf_counter() - start) * 1000)
    if replicas and resp(PORT, "WAIT", "1", "5000") != "1":
        raise RuntimeError("La réplica no confirmó el lote de escrituras")
    count = int(resp(PORT, "LLEN", key))
    replica_count = int(resp(REPLICA_PORT, "LLEN", key)) if replicas else None
    if count != SAMPLES or (replicas and replica_count != SAMPLES):
        raise RuntimeError(f"Datos incompletos: {count}, réplica: {replica_count}")
    ordered = sorted(times)
    return {"scenario": name, "writes": SAMPLES, "primary_count": count,
            "replica_count": replica_count, "mean_ms": round(statistics.mean(times), 3),
            "p95_ms": round(ordered[int(SAMPLES * .95) - 1], 3),
            "writes_per_second": round(SAMPLES / (sum(times) / 1000), 2)}


def main():
    for name in (REPLICA, PRIMARY):
        subprocess.run(["docker", "rm", "-f", name], capture_output=True)
    subprocess.run(["docker", "network", "rm", NETWORK], capture_output=True)
    try:
        docker("network", "create", NETWORK)
        docker("run", "-d", "--rm", "--name", PRIMARY, "--network", NETWORK,
               "-p", f"127.0.0.1:{PORT}:6379", IMAGE,
               "valkey-server", "--save", "", "--appendonly", "no")
        ready(PORT)
        one = trial("1 primario", False)
        docker("run", "-d", "--rm", "--name", REPLICA, "--network", NETWORK,
               "-p", f"127.0.0.1:{REPLICA_PORT}:6379", IMAGE,
               "valkey-server", "--replicaof", PRIMARY, "6379",
               "--save", "", "--appendonly", "no")
        ready(REPLICA_PORT)
        for _ in range(200):
            if "master_link_status:up" in resp(REPLICA_PORT, "INFO", "replication"):
                break
            time.sleep(0.1)
        else:
            raise RuntimeError("Replicación no establecida")
        two = trial("1 primario + 1 réplica", True)
        print(json.dumps({"method": "1000 RPUSH secuenciales por escenario; cada conexión es nueva;"
                           " WAIT 1 se ejecuta después de medir el segundo lote", "results": [one, two]},
                         ensure_ascii=False, indent=2))
    finally:
        for name in (REPLICA, PRIMARY):
            subprocess.run(["docker", "rm", "-f", name], capture_output=True)
        subprocess.run(["docker", "network", "rm", NETWORK], capture_output=True)


if __name__ == "__main__":
    main()
