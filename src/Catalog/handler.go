package Catalog

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	log *zap.Logger
}

func NewHandler(s Service, log *zap.Logger) *Handler {
	return &Handler{service: s, log: log}
}

// ---------------- CATEGORY HANDLERS -----------------

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Creates a category with name, slug, description and optional parent
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      CreateCategoryRequest  true  "Category payload"
// @Success      201       {object}  CategoryResponse
// @Failure      400       {object}  map[string]interface{}
// @Failure      422       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /categories [post]
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var dto CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	category, err := h.service.CreateCategory(r.Context(), dto)
	if err != nil {
		h.handleServiceError(w, err, "create category")
		return
	}

	h.writeJSON(w, http.StatusCreated, category)
}

// GetCategory godoc
// @Summary      Get category by ID
// @Description  Returns a single category by its UUID
// @Tags         categories
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      200  {object}  CategoryResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categories/{id} [get]
func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid category id format")
		return
	}

	category, err := h.service.GetCategory(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err, "get category")
		return
	}

	h.writeJSON(w, http.StatusOK, category)
}

// GetCategoryBySlug godoc
// @Summary      Get category by slug
// @Description  Returns a single category by its slug
// @Tags         categories
// @Produce      json
// @Param        slug   path      string  true  "Category slug"
// @Success      200  {object}  CategoryResponse
// @Failure      404  {object}  map[string]interface{}
// @Router       /categories/slug/{slug} [get]
func (h *Handler) GetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		h.writeError(w, http.StatusBadRequest, "slug is required")
		return
	}

	category, err := h.service.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		h.handleServiceError(w, err, "get category by slug")
		return
	}

	h.writeJSON(w, http.StatusOK, category)
}

// UpdateCategory godoc
// @Summary      Update category
// @Description  Updates a category by ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id        path      string                 true  "Category ID"
// @Param        category  body      UpdateCategoryRequest  true  "Category payload"
// @Success      200       {object}  CategoryResponse
// @Failure      400       {object}  map[string]interface{}
// @Failure      404       {object}  map[string]interface{}
// @Failure      409       {object}  map[string]interface{}
// @Router       /categories/{id} [put]
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid category id format")
		return
	}

	var dto UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	category, err := h.service.UpdateCategory(r.Context(), id, dto)
	if err != nil {
		h.handleServiceError(w, err, "update category")
		return
	}

	h.writeJSON(w, http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary      Delete category
// @Description  Deletes a category by ID
// @Tags         categories
// @Param        id   path      string  true  "Category ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categories/{id} [delete]
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid category id format")
		return
	}

	if err := h.service.DeleteCategory(r.Context(), id); err != nil {
		h.handleServiceError(w, err, "delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListCategories godoc
// @Summary      List categories
// @Description  Returns a paginated list of categories with optional filtering
// @Tags         categories
// @Produce      json
// @Param        limit     query     int     false  "Number of items to return (max 100)"  default(20)
// @Param        offset    query     int     false  "Number of items to skip"              default(0)
// @Param        search    query     string  false  "Search term for name or slug"
// @Param        parent_id query     string  false  "Parent category UUID"
// @Success      200       {object}  PaginatedResponse
// @Failure      400       {object}  map[string]interface{}
// @Router       /categories [get]
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	q := ListCategoriesQuery{Limit: 20}

	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil && limit > 0 && limit <= 100 {
			q.Limit = limit
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil && offset >= 0 {
			q.Offset = offset
		}
	}

	if s := r.URL.Query().Get("search"); s != "" {
		q.Search = s
	}

	if p := r.URL.Query().Get("parent_id"); p != "" {
		if parentID, err := uuid.Parse(p); err == nil {
			q.ParentID = &parentID
		}
	}

	result, err := h.service.ListCategories(r.Context(), q)
	if err != nil {
		h.handleServiceError(w, err, "list categories")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// ---------------- PRODUCT HANDLERS -----------------

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Creates a product with name, description, price, currency and optional category
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        product  body      CreateProductRequest  true  "Product payload"
// @Success      201      {object}  ProductResponse
// @Failure      400      {object}  map[string]interface{}
// @Failure      422      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /products [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var dto CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), dto)
	if err != nil {
		h.handleServiceError(w, err, "create product")
		return
	}

	h.writeJSON(w, http.StatusCreated, product)
}

// GetProduct godoc
// @Summary      Get product by ID
// @Description  Returns a single product by its UUID
// @Tags         products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  ProductResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err, "get product")
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// GetProductBySKU godoc
// @Summary      Get product by SKU
// @Description  Returns a single product by its SKU
// @Tags         products
// @Produce      json
// @Param        sku   path      string  true  "Product SKU"
// @Success      200  {object}  ProductResponse
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/sku/{sku} [get]
func (h *Handler) GetProductBySKU(w http.ResponseWriter, r *http.Request) {
	sku := chi.URLParam(r, "sku")
	if sku == "" {
		h.writeError(w, http.StatusBadRequest, "SKU is required")
		return
	}

	product, err := h.service.GetProductBySKU(r.Context(), sku)
	if err != nil {
		h.handleServiceError(w, err, "get product by SKU")
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary      Update product
// @Description  Updates a product by ID
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        id       path      string               true  "Product ID"
// @Param        product  body      UpdateProductRequest true  "Product payload"
// @Success      200      {object}  ProductResponse
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Router       /products/{id} [put]
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	var dto UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, dto)
	if err != nil {
		h.handleServiceError(w, err, "update product")
		return
	}

	h.writeJSON(w, http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary      Delete product
// @Description  Deletes a product by ID
// @Tags         products
// @Param        id   path      string  true  "Product ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [delete]
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		h.handleServiceError(w, err, "delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProducts godoc
// @Summary      List products
// @Description  Returns a paginated list of products with optional filtering
// @Tags         products
// @Produce      json
// @Param        limit       query     int     false  "Number of items to return (max 100)"  default(20)
// @Param        offset      query     int     false  "Number of items to skip"              default(0)
// @Param        search      query     string  false  "Search term for name or SKU"
// @Param        category_id query     string  false  "Category UUID"
// @Param        min_price   query     number  false  "Minimum price filter"
// @Param        max_price   query     number  false  "Maximum price filter"
// @Success      200         {object}  PaginatedResponse
// @Failure      400         {object}  map[string]interface{}
// @Router       /products [get]
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := ListProductsQuery{Limit: 20}

	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil && limit > 0 && limit <= 100 {
			q.Limit = limit
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil && offset >= 0 {
			q.Offset = offset
		}
	}

	if s := r.URL.Query().Get("search"); s != "" {
		q.Search = s
	}

	if c := r.URL.Query().Get("category_id"); c != "" {
		if categoryID, err := uuid.Parse(c); err == nil {
			q.CategoryID = &categoryID
		}
	}

	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			decimal := decimal.NewFromFloat(minPrice)
			q.MinPrice = &decimal
		}
	}

	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			decimal := decimal.NewFromFloat(maxPrice)
			q.MaxPrice = &decimal
		}
	}

	result, err := h.service.ListProducts(r.Context(), q)
	if err != nil {
		h.handleServiceError(w, err, "list products")
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// ---------------- UTILITY METHODS -----------------

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error":     msg,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    status,
	})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error, operation string) {
	h.log.Error(operation+" failed", zap.Error(err))

	switch err {
	case ErrProductNotFound, ErrCategoryNotFound:
		h.writeError(w, http.StatusNotFound, err.Error())
	case ErrProductConflict, ErrCategoryConflict, ErrProductSKUExists, ErrCategorySlugExists:
		h.writeError(w, http.StatusConflict, err.Error())
	case ErrCategoryCircularRef:
		h.writeError(w, http.StatusBadRequest, err.Error())
	case ErrValidationFailed:
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// RegisterRoutes sets up all the routes for the catalog module
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/categories", func(r chi.Router) {
		r.Post("/", h.CreateCategory)
		r.Get("/", h.ListCategories)
		r.Get("/{id}", h.GetCategory)
		r.Put("/{id}", h.UpdateCategory)
		r.Delete("/{id}", h.DeleteCategory)
		r.Get("/slug/{slug}", h.GetCategoryBySlug)
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.ListProducts)
		r.Get("/{id}", h.GetProduct)
		r.Put("/{id}", h.UpdateProduct)
		r.Delete("/{id}", h.DeleteProduct)
		r.Get("/sku/{sku}", h.GetProductBySKU)
	})
}
