package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/al-mighty/linkpulse/internal/models"
	"github.com/al-mighty/linkpulse/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc    *service.LinkService
	events *service.EventService
}

func New(svc *service.LinkService, events *service.EventService) *Handler {
	return &Handler{svc: svc, events: events}
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

// TrackEvent: POST /api/events  body: {project, name, page?, payload?}
// Fast path — enqueues to a channel and returns 202.
func (h *Handler) TrackEvent(w http.ResponseWriter, r *http.Request) {
	var req models.TrackEventReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	ev := &models.Event{
		Project:   req.Project,
		Name:      req.Name,
		Page:      req.Page,
		Payload:   req.Payload,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
	}
	if err := h.events.Track(ev); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"ok"}`))
}

// EventStats: GET /api/events/stats?project=X&since=30
func (h *Handler) EventStats(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	sinceDays, _ := strconv.Atoi(r.URL.Query().Get("since"))
	stats, err := h.events.Stats(r.Context(), project, sinceDays)
	if err != nil {
		http.Error(w, `{"error":"failed to read stats"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}