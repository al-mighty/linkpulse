package handler

import (
	"encoding/json"
	"net/http"

	"github.com/al-mighty/linkpulse/internal/models"
	"github.com/al-mighty/linkpulse/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.LinkService
}

func New(svc *service.LinkService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Create(r.Context(), req.URL)
	if err != nil {
		http.Error(w, `{"error":"failed to create link"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	url, linkID, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Track click async
	h.svc.TrackClick(linkID, r.RemoteAddr, r.UserAgent(), r.Referer())

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	links, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list"}`, http.StatusInternalServerError)
		return
	}
	if links == nil {
		links = []models.Link{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	stats, err := h.svc.Stats(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}