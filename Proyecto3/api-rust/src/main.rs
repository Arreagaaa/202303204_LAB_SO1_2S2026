use axum::{extract::State, http::StatusCode, routing::{get, post}, Json, Router};
use chrono::DateTime;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::{env, net::SocketAddr};

#[derive(Clone)]
struct AppState {
    ingest_url: String,
    client: Client,
}

#[derive(Debug, Deserialize, Serialize)]
struct Report {
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: String,
}

fn valid_report(report: &Report) -> bool {
    matches!(report.country.as_str(), "USA" | "RUS" | "CHN" | "ESP" | "GTM")
        && (0..=50).contains(&report.warplanes_in_air)
        && (0..=30).contains(&report.warships_in_water)
        && DateTime::parse_from_rfc3339(&report.timestamp).is_ok()
}

async fn health() -> &'static str {
    "ok"
}

async fn receive_report(
    State(state): State<AppState>,
    Json(report): Json<Report>,
) -> Result<(StatusCode, Json<serde_json::Value>), (StatusCode, String)> {
    if !valid_report(&report) {
        return Err((StatusCode::BAD_REQUEST, "reporte fuera de rango".to_string()));
    }

    let response = state.client.post(&state.ingest_url).json(&report).send().await
        .map_err(|error| (StatusCode::BAD_GATEWAY, format!("servicio Go no disponible: {error}")))?;
    if !response.status().is_success() {
        return Err((StatusCode::BAD_GATEWAY, "el servicio Go rechazo el reporte".to_string()));
    }
    Ok((StatusCode::ACCEPTED, Json(serde_json::json!({"status": "accepted"}))))
}

#[tokio::main]
async fn main() {
    let ingest_url = env::var("INGEST_URL")
        .unwrap_or_else(|_| "http://go-ingest:8081/reports".to_string());
    let state = AppState { ingest_url, client: Client::new() };
    let app = Router::new()
        .route("/", get(health))
        .route("/health", get(health))
        .route("/grpc-2023204", post(receive_report))
        .with_state(state);
    let address = SocketAddr::from(([0, 0, 0, 0], 8080));
    println!("api-rust escuchando en {address}");
    let listener = tokio::net::TcpListener::bind(address).await.expect("no se pudo abrir el puerto");
    axum::serve(listener, app).await.expect("fallo del servidor");
}

#[cfg(test)]
mod tests {
    use super::{valid_report, Report};

    fn report() -> Report {
        Report {
            country: "GTM".to_string(),
            warplanes_in_air: 20,
            warships_in_water: 10,
            timestamp: "2026-10-07T23:40:00Z".to_string(),
        }
    }

    #[test]
    fn acepta_reporte_valido() {
        assert!(valid_report(&report()));
    }

    #[test]
    fn rechaza_pais_no_definido() {
        let mut value = report();
        value.country = "MEX".to_string();
        assert!(!valid_report(&value));
    }

    #[test]
    fn rechaza_valores_fuera_de_rango() {
        let mut value = report();
        value.warplanes_in_air = 51;
        assert!(!valid_report(&value));
    }

    #[test]
    fn rechaza_timestamp_invalido() {
        let mut value = report();
        value.timestamp = "ahora".to_string();
        assert!(!valid_report(&value));
    }
}
