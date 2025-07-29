package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"officestonks/internal/services"
)

type NewsHandler struct {
	service *services.NewsService
}

func NewNewsHandler(service *services.NewsService) *NewsHandler {
	return &NewsHandler{service: service}
}

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Expires string `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}
	if req.Expires == "" {
		http.Error(w, "Expiration date is required", http.StatusBadRequest)
		return
	}
	
	exp, err := time.Parse(time.RFC3339, req.Expires)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid expiration format (expected RFC3339): %v", err), http.StatusBadRequest)
		return
	}
	if err := h.service.CreateNews(req.Title, req.Content, exp); err != nil {
		http.Error(w, "Failed to create news", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *NewsHandler) GetActiveNews(w http.ResponseWriter, r *http.Request) {
	news, err := h.service.GetActiveNews()
	if err != nil {
		http.Error(w, "Failed to fetch news", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(news)
}
