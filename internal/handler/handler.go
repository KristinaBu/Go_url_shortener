package handler

import (
	"encoding/json"
	"errors"
	"github.com/KristinaBu/Go_url_shortener/internal/domain"
	"github.com/KristinaBu/Go_url_shortener/internal/service"
	"io"
	"net/http"
	"strings"
)

const prefix = "/links/"

type Handler struct {
	service *service.LinkService
}

func New(svc *service.LinkService) *Handler {
	return &Handler{
		service: svc,
	}
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	ShortURL string `json:"short_url"`
}

type getLinkResponse struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request createLinkRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	link, err := h.service.Create(r.Context(), request.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidURL):
			writeError(w, http.StatusBadRequest, "invalid URL")

		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	response := createLinkResponse{
		ShortURL: prefix + link.ShortCode,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, prefix)

	if shortCode == "" {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}

	link, err := h.service.Get(r.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, "link not found")

		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	response := getLinkResponse{
		URL: link.OriginalURL,
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Error: message,
	})
}
