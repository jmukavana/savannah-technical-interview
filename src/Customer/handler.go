package Customer

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	svc Service
	log *zap.Logger
	v   *validator.Validate
}

func NewHandler(s Service, log *zap.Logger) *Handler {
	v := validator.New()
	_ = v.RegisterValidation("e164", func(fl validator.FieldLevel) bool {
		// simple e164 check: starts with + and digits, length 7-15
		s := fl.Field().String()
		if len(s) < 7 || len(s) > 15 { // includes +
			return false
		}
		if s[0] != '+' {
			return false
		}
		for i := 1; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return false
			}
		}
		return true
	})
	return &Handler{svc: s, log: log, v: v}
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := ListCustomersQuery{Limit: 20}
	// parse query params
	if l := r.URL.Query().Get("limit"); l != "" {
		// ignore parse errors for brevity; production: validate properly
	}
	if s := r.URL.Query().Get("search"); s != "" {
		q.Search = s
	}
	if st := r.URL.Query().Get("status"); st != "" {
		q.Status = st
	}
	customers, err := h.svc.List(r.Context(), q)
	if err != nil {
		h.log.Error("list customers", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to list")
		return
	}
	h.writeJSON(w, http.StatusOK, customers)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var dto CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.v.Struct(dto); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	c, err := h.svc.Create(r.Context(), dto)
	if err != nil {
		h.log.Error("create", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create")
		return
	}
	h.writeJSON(w, http.StatusCreated, c)
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if err == ErrorNotFound {
			h.writeError(w, http.StatusNotFound, "not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed")
		return
	}
	h.writeJSON(w, http.StatusOK, c)
	if err == ErrorNotFound {
		h.writeError(w, http.StatusNotFound, "not found")
		return
	}
	h.writeError(w, http.StatusInternalServerError, "failed to update")
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var dto UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.v.Struct(dto); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.svc.Update(r.Context(), id, dto)
	if err != nil {
		if err == ErrorConflict {
			h.writeError(w, http.StatusConflict, "version conflict")
			return
		}
		if err == ErrorNotFound {
			h.writeError(w, http.StatusNotFound, "not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "failed to update")
		return
	}
	h.writeJSON(w, http.StatusOK, updated)
}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]interface{}{"error": msg, "timestamp": time.Now().UTC()})
}
