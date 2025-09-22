package engine

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/KRAZYFLASH/carZone/models"
	"github.com/KRAZYFLASH/carZone/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
)

type EngineHandler struct {
	service service.EngineServiceInterface
}

func NewEngineHandler(service service.EngineServiceInterface) *EngineHandler {
	return &EngineHandler{service: service}
}

func (h *EngineHandler) GetEngineById(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("EngineHandler")
	ctx, span := tracer.Start(r.Context(), "GetEngineById-Handler")
	defer span.End()

	vars := mux.Vars(r)
	id := vars["id"]

	resp, err := h.service.GetEngineById(ctx, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error fetching engine by ID:", err)
		return
	}

	body, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error marshalling response:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write the response body
	_, err = w.Write(body)
	if err != nil {
		log.Println("Error writing response:", err)
	}

}

func (h *EngineHandler) CreateEngine(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("EngineHandler")
	ctx, span := tracer.Start(r.Context(), "CreateEngine-Handler")
	defer span.End()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error reading request body:", err)
		return
	}

	var engineReq models.EngineRequest

	err = json.Unmarshal(body, &engineReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error unmarshalling request body:", err)
		return	
	}

	createdEngine, err := h.service.CreateEngine(ctx, &engineReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error creating engine:", err)
		return
	}

	responseBody, err := json.Marshal(createdEngine)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error marshalling response:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(responseBody)

	if err != nil {
		log.Println("Error writing response:", err)
	}

}

func (h *EngineHandler) UpdateEngine(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("EngineHandler")
	ctx, span := tracer.Start(r.Context(), "UpdateEngine-Handler")
	defer span.End()

	vars := mux.Vars(r)
	id := vars["id"]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error reading request body:", err)
		return
	}

	var engineReq models.EngineRequest

	err = json.Unmarshal(body, &engineReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error unmarshalling request body:", err)
		return	
	}

	updatedEngine, err := h.service.UpdateEngine(ctx, id, &engineReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error updating engine:", err)
		return
	}	

	responseBody, err := json.Marshal(updatedEngine)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error marshalling response:", err)
		return
	}	

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseBody)
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func (h *EngineHandler) DeleteEngine(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("EngineHandler")
	ctx, span := tracer.Start(r.Context(), "DeleteEngine-Handler")
	defer span.End()
	
	vars := mux.Vars(r)
	id := vars["id"]

	deletedEngine, err := h.service.DeleteEngine(ctx, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error deleting engine:", err)
		response := map[string]string{"error": "Error deleting engine"}
		responseBody, _ := json.Marshal(response)
		_, _ = w.Write(responseBody)
		return
	}

	if deletedEngine.EngineID == uuid.Nil{
		w.WriteHeader(http.StatusNotFound)
		responseBody := map[string]string{"error": "Engine not found"}
		jsonResponse, _ := json.Marshal(responseBody)
		_, _ = w.Write(jsonResponse)
		return
	}

	jsonResponse, err := json.Marshal(deletedEngine)
	if err != nil {
		log.Println("Error marshalling response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		responseBody := map[string]string{"error": "Error processing response"}
		jsonResponse, _ := json.Marshal(responseBody)
		_, _ = w.Write(jsonResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)

}