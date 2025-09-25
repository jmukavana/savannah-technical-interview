package Orders

import (
	
	"net/http"
	

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service   Service
	validator *validator.Validate
	logger    *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(),
		logger:    logger,
	}
}

// CreateOrder handles order creation
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder handles retrieving a single order
func (h *Handler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrders handles listing orders with filtering and pagination
func (h *Handler) ListOrders(c *gin.Context) {
	var req OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	orders, err := h.service.ListOrders(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus handles updating order status
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	if err := h.service.UpdateOrderStatus(c.Request.Context(), id, req.Status, req.Version); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}

// CancelOrder handles order cancellation
func (h *Handler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	if err := h.service.CancelOrder(c.Request.Context(), id, req.Version, req.Reason); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}

// GetOrderItems handles retrieving order items
func (h *Handler) GetOrderItems(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	items, err := h.service.GetOrderItems(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// handleError handles different types of errors appropriately
func (h *Handler) handleError(c *gin.Context, err error) {
	h.logger.Error("handler error", zap.Error(err))

	if orderErr, ok := err.(OrderError); ok {
		c.JSON(orderErr.Status, gin.H{"error": orderErr.Message, "code": orderErr.Code})
		return
	}

	switch err {
	case ErrOrderNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
	case ErrVersionConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "version conflict"})
	case ErrInvalidOrderStatus:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order status"})
	case ErrInsufficientInventory:
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient inventory"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// RegisterRoutes registers all order-related routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("", h.ListOrders)
		orders.GET("/:id", h.GetOrder)
		orders.PATCH("/:id/status", h.UpdateOrderStatus)
		orders.POST("/:id/cancel", h.CancelOrder)
		orders.GET("/:id/items", h.GetOrderItems)
	}
}