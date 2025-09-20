package Catalog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	log *zap.Logger
}

func NewHandler(s Service, log *zap.Logger) *Handler {
	return &Handler{service: s, log: log}
}

// ---------------- CATEGORY -----------------

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Creates a category with name, slug, description
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      CreateCategoryRequest  true  "Category payload"
// @Success      201       {object}  Category
// @Failure      400       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /categories [post]

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var dto CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	c, err := h.service.CreateCategory(r.Context(), dto)
	if err != nil {
		h.log.Error("create category", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}
	h.writeJSON(w, http.StatusCreated, c)
}
// GetProduct godoc
// @Summary      Get category by ID
// @Description  Returns a single category by its UUID
// @Tags         products
// @Produce      json
// @Param        id   path      string  true  "ID"
// @Success      200  {object}  Category
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [get]

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	c, err := h.service.GetCategory(r.Context(), id)
	if err != nil {
		if err == CategoryErrorNotFound {
			h.writeError(w, http.StatusNotFound, "category not found")
			return
		}
		h.log.Error("get category", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get category")
		return
	}
	h.writeJSON(w, http.StatusOK, c)
}

// ---------------- PRODUCT -----------------

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var dto CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	p, err := h.service.CreateProduct(r.Context(), dto)
	if err != nil {
		h.log.Error("create product", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to create product")
		return
	}
	h.writeJSON(w, http.StatusCreated, p)
}
// GetProduct godoc
// @Summary      Get product by ID
// @Description  Returns a single product by its UUID
// @Tags         products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if err == ProductErrorNotFound {
			h.writeError(w, http.StatusNotFound, "product not found")
			return
		}
		h.log.Error("get product", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "failed to get product")
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
    q := ListProductsQuery{Limit: 20} // default

    if l := r.URL.Query().Get("limit"); l != "" {
        if limit, err := strconv.Atoi(l); err == nil {
            if limit > 0 && limit <= 100 {
                q.Limit = limit
            }
        }
    }

    if s := r.URL.Query().Get("search"); s != "" {
        q.Search = s
    }

    products, err := h.service.ListProducts(r.Context(), q)
    if err != nil {
        h.log.Error("list products", zap.Error(err))
        h.writeError(w, http.StatusInternalServerError, "failed to list products")
        return
    }
    h.writeJSON(w, http.StatusOK, products)
}

// ---------------- UTIL -----------------

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error":     msg,
		"timestamp": time.Now().UTC(),
	})
}
