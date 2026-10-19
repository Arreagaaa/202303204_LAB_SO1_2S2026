"""Generate the provisioned Grafana dashboard from Prometheus metrics."""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "operator/helm-charts/grafanainstance/dashboards/reports.json"


def panel(number, title, expression, kind, x, y, width=6, height=5):
    result = {
        "id": number,
        "type": kind,
        "title": title,
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "gridPos": {"x": x, "y": y, "w": width, "h": height},
        "targets": [{"refId": "A", "expr": expression, "instant": kind != "timeseries"}],
    }
    if kind == "bargauge":
        result["options"] = {"orientation": "horizontal", "displayMode": "gradient"}
    return result


dashboard = {
    "uid": "mumnk8s-reports",
    "title": "M.U.M.N.K8s - Reportes militares",
    "schemaVersion": 39,
    "version": 1,
    "refresh": "10s",
    "tags": ["mumnk8s", "202303204"],
    "panels": [
        panel(1, "Maximo de aviones en aire", "military_warplanes_max", "stat", 0, 0),
        panel(2, "Minimo de aviones en aire", "military_warplanes_min", "stat", 6, 0),
        panel(3, "Maximo de barcos en mar", "military_warships_max", "stat", 12, 0),
        panel(4, "Minimo de barcos en mar", "military_warships_min", "stat", 18, 0),
        panel(5, "Top paises por aviones", "topk(5, military_country_warplanes_total)", "bargauge", 0, 5, 12, 7),
        panel(6, "Top paises por barcos", "topk(5, military_country_warships_total)", "bargauge", 12, 5, 12, 7),
        panel(7, "Moda de aviones", "military_warplanes_mode", "stat", 0, 12),
        panel(8, "Moda de barcos", "military_warships_mode", "stat", 6, 12),
        {
            "id": 9,
            "type": "text",
            "title": "Pais asignado por carne 202303204",
            "gridPos": {"x": 12, "y": 12, "w": 6, "h": 5},
            "options": {"mode": "markdown", "content": "# CHN\nUltimo digito del carne: 4"},
        },
        panel(10, "Reportes recibidos de CHN", 'military_country_reports_total{country="CHN"}', "stat", 18, 12),
        panel(11, "CHN: aviones en aire", "military_assigned_warplanes", "timeseries", 0, 17, 12, 8),
        panel(12, "CHN: barcos en mar", "military_assigned_warships", "timeseries", 12, 17, 12, 8),
    ],
}

OUTPUT.write_text(json.dumps(dashboard, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
