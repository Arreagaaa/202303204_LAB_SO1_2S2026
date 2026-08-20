package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Configuracion de esta API.
// Corre en VM2 (podman) y escucha en el puerto 8083.
// Esta API se comunica con API1 y API2 (ambas en VM1).
const (
	apiName = "API3"
	vmNum   = 2
	carnet  = "202303204"
	port    = ":8083"
)

// HealthResponse es el formato JSON que se retorna al consultar /health.
// Lo usan las otras APIs para verificar si esta API esta operativa.
type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	VM        int    `json:"VM"`
	Carnet    string `json:"carnet"`
}

// CallResponse es el formato JSON que se retorna al consultar un endpoint call-api.
// Indica si la API destino esta respondiendo correctamente o no.
type CallResponse struct {
	ApiName    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     string `json:"carnet"`
}

// healthHandler responde a GET /health con el estado de esta API.
// Las demas APIs llaman a este endpoint para saber si API3 esta arriba.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "UP",
		Message:   apiName + " is Ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		VM:        vmNum,
		Carnet:    carnet,
	})
}

// callOtherAPI hace un GET al /health de otra API y retorna un CallResponse.
// Si la API destino no responde o no retorna status "UP", retorna connection: false.
// Ejemplo: API3 llama a API1 (VM1, 192.0.2.11:8081) o a API2 (VM1, 192.0.2.11:8082).
func callOtherAPI(targetName string, targetVM int, targetURL string) CallResponse {
	resp, err := http.Get(targetURL + "/health")
	if err != nil {
		return CallResponse{
			ApiName:    targetName,
			Message:    fmt.Sprintf("ERROR: The %s located on the VM%d is not working", targetName, targetVM),
			Connection: false,
			Carnet:     carnet,
		}
	}
	defer resp.Body.Close()

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return CallResponse{
			ApiName:    targetName,
			Message:    fmt.Sprintf("ERROR: The %s located on the VM%d is not working", targetName, targetVM),
			Connection: false,
			Carnet:     carnet,
		}
	}

	if health.Status == "UP" {
		return CallResponse{
			ApiName:    targetName,
			Message:    fmt.Sprintf("The %s located on the VM%d is working", targetName, targetVM),
			Connection: true,
			Carnet:     carnet,
		}
	}

	return CallResponse{
		ApiName:    targetName,
		Message:    fmt.Sprintf("ERROR: The %s located on the VM%d is not working", targetName, targetVM),
		Connection: false,
		Carnet:     carnet,
	}
}

// writeJSON escribe una respuesta JSON en el ResponseWriter.
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getEnv lee una variable de entorno. Si no existe, retorna el valor por defecto.
// Se usa para configurar las URLs de las otras APIs sin hardcodear en el codigo.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	// URLs de las otras APIs. Se pueden sobreescribir con variables de entorno.
	// API1 y API2 estan en VM1.
	api1URL := getEnv("API1_URL", "http://192.0.2.11:8081")
	api2URL := getEnv("API2_URL", "http://192.0.2.11:8082")

	// Endpoint de salud. Cada API debe tener este endpoint.
	http.HandleFunc("/health", healthHandler)

	// GET /api3/202303204/call-api1 -> llama al /health de API1 (VM1, puerto 8081).
	http.HandleFunc("/api3/"+carnet+"/call-api1", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, callOtherAPI("API1", 1, api1URL))
	})

	// GET /api3/202303204/call-api2 -> llama al /health de API2 (VM1, puerto 8082).
	http.HandleFunc("/api3/"+carnet+"/call-api2", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, callOtherAPI("API2", 1, api2URL))
	})

	fmt.Println("API3 corriendo en puerto " + port)
	http.ListenAndServe(port, nil)
}
