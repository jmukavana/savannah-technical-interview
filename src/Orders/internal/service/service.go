package Orders

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"	
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel"	
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	

	
)

// ServiceConfig holds dependencies for the service.
type ServiceConfig struct {
	Repository      Repository
	DB              *sqlx.DB
	Inventory       InventoryService
	TaxCalculator   TaxCalculator
	ShippingCalc    ShippingCalculator
	Notification    NotificationService
	Payment         PaymentService
	Audit           AuditService
	Logger          *zap.Logger
	Tracer          trace.Tracer
}

// service implements the Service interface.
type service struct {
	repo         Repository
	db           *sqlx.DB
	inv          InventoryService
	taxCalc      TaxCalculator
	shippingCalc ShippingCalculator
	notification NotificationService
	payment      PaymentService
	audit        AuditService
	log          *zap.Logger
	tracer       trace.Tracer
}

// NewService creates a new order service.
func NewService(cfg ServiceConfig) Service {
	tracer := cfg.Tracer
	if tracer == nil {
		tracer = otel.Tracer("orders-service")
	}
	return &service{
		repo:         cfg.Repository,
		db:           cfg.DB,
		inv:          cfg.Inventory,
		taxCalc:      cfg.TaxCalculator,
		shippingCalc: cfg.ShippingCalc,
		notification: cfg.Notification,
		payment:      cfg.Payment,
		audit:        cfg.Audit,
		log:          cfg.Logger,
		tracer:       tracer,
	}
}

// CreateOrder creates a new order with the given request.
func (s *service) CreateOrder(ctx context.Context, req CreateOrderRequest) (*model.OrderResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.CreateOrder")
	defer span.End()

	// Validate input
	if err := s.validateOrderData(ctx, req); err != nil {
		metrics.OrderCreationErrors.Inc()
		return nil, errors.Wrap(err, "validation failed")
	}

	// Convert and calculate items
	orderItems, subtotal, err := s.processOrderItems(ctx, req.Items)
	if err != nil {
		metrics.OrderCreationErrors.Inc()
		return nil, errors.Wrap(err, "failed to process order items")
	}

	// Calculate tax
	tax, err := s.calculateTax(ctx, subtotal, req.CustomerID)
	if err != nil {
		s.log.Error("failed to calculate tax", zap.Error(err))
		tax = subtotal.Mul(decimal.NewFromFloat(0.08)) // Fallback
	}

	// Calculate shipping
	shipping, err := s.calculateShipping(ctx, req.Items, req.CustomerID)
	if err != nil {
		s.log.Error("failed to calculate shipping", zap.Error(err))
		shipping = decimal.NewFromFloat(9.99) // Fallback
	}

	total := subtotal.Add(tax).Add(shipping)
	order := &model.Order{
		CustomerID: req.CustomerID,
		Status:     model.StatusCreated,
		Subtotal:   subtotal,
		Tax:        tax,
		Shipping:   shipping,
		Total:      total,
		Currency:   req.Currency,
		Warehouse:  req.Warehouse,
		Version:    1,
		Metadata:   req.Metadata,
	}

	// Begin transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		metrics.OrderCreationErrors.Inc()
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Reserve inventory
	for _, item := range orderItems {
		if item.ProductID != nil {
			if err := s.inv.Reserve(ctx, *item.ProductID, item.Quantity, req.Warehouse); err != nil {
				s.log.Error("inventory reservation failed",
					zap.String("product_id", item.ProductID.String()),
					zap.Int("quantity", item.Quantity),
					zap.Error(err))
				metrics.OrderCreationErrors.Inc()
				return nil, ErrInsufficientInventory
			}
		}
	}

	// Save order and items
	if err = s.repo.CreateOrderTx(ctx, tx, order, orderItems, req.Addresses); err != nil {
		metrics.OrderCreationErrors.Inc()
		return nil, errors.Wrap(err, "failed to create order")
	}

	// Process payment
	if s.payment != nil {
		if err := s.payment.ProcessPayment(ctx, order.ID, order.Total); err != nil {
			s.log.Error("payment processing failed", zap.String("order_id", order.ID.String()), zap.Error(err))
			metrics.OrderCreationErrors.Inc()
			return nil, errors.Wrap(err, "payment processing failed")
		}
	}

	if err = tx.Commit(); err != nil {
		metrics.OrderCreationErrors.Inc()
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	// Send notifications and audit events asynchronously
	s.notifyAndAuditOrderCreated(ctx, order, orderItems, req.CustomerID)

	metrics.OrderCreations.Inc()
	return s.orderToResponse(order, orderItems, req.Addresses), nil
}

// GetOrderByID retrieves an order by its ID.
func (s *service) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.OrderResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetOrderByID")
	defer span.End()

	order, items, addresses, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFoundWithID(id)
		}
		return nil, errors.Wrap(err, "failed to get order")
	}
	return s.orderToResponse(order, items, addresses), nil
}

// ListOrders lists orders with pagination and filtering.
func (s *service) ListOrders(ctx context.Context, req *model.OrderListRequest) (*model.OrderListResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.ListOrders")
	defer span.End()

	orders, total, err := s.repo.ListOrders(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list orders")
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		_, items, addresses, err := s.repo.GetOrder(ctx, order.ID)
		if err != nil {
			s.log.Error("failed to get order items", zap.String("order_id", order.ID.String()), zap.Error(err))
			items = []model.OrderItem{}
			addresses = []model.OrderAddress{}
		}
		responses[i] = *s.orderToResponse(&order, items, addresses)
	}

	return &model.OrderListResponse{
		Orders: responses,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// UpdateOrderStatus updates the status of an order.
func (s *service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, version int) error {
	ctx, span := s.tracer.Start(ctx, "service.UpdateOrderStatus")
	defer span.End()

	order, _, _, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFoundWithID(id)
		}
		return errors.Wrap(err, "failed to get order")
	}

	if !order.CanTransitionTo(status) {
		return ErrInvalidOrderStatusTransition(order.Status, status)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = s.repo.UpdateOrderStatusTx(ctx, tx, id, status, version); err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			return ErrVersionConflict
		}
		return errors.Wrap(err, "failed to update order status")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	s.notifyAndAuditStatusUpdate(ctx, id, status, order.Status, version)
	metrics.OrderStatusUpdates.WithLabelValues(status).Inc()
	return nil
}

// CancelOrder cancels an order with a reason.
func (s *service) CancelOrder(ctx context.Context, id uuid.UUID, version int, reason string) error {
	ctx, span := s.tracer.Start(ctx, "service.CancelOrder")
	defer span.End()

	order, items, _, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFoundWithID(id)
		}
		return errors.Wrap(err, "failed to get order")
	}

	if !order.IsCancellable() {
		return ErrCannotCancelOrder
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, item := range items {
		if item.ProductID != nil {
			if err := s.inv.Release(ctx, *item.ProductID, item.Quantity, order.Warehouse); err != nil {
				s.log.Error("failed to release inventory",
					zap.String("product_id", item.ProductID.String()),
					zap.Int("quantity", item.Quantity),
					zap.Error(err))
			}
		}
	}

	if err = s.repo.UpdateOrderStatusTx(ctx, tx, id, model.StatusCancelled, version); err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			return ErrVersionConflict
		}
		return errors.Wrap(err, "failed to cancel order")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	s.notifyAndAuditCancellation(ctx, order, reason, version)
	metrics.OrderCancellations.Inc()
	return nil
}

// RefundOrder processes a refund for an order.
func (s *service) RefundOrder(ctx context.Context, id uuid.UUID, version int, reason string) error {
	ctx, span := s.tracer.Start(ctx, "service.RefundOrder")
	defer span.End()

	order, _, _, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFoundWithID(id)
		}
		return errors.Wrap(err, "failed to get order")
	}

	if !order.IsRefundable() {
		return ErrInvalidOrderStatusTransition(order.Status, model.StatusRefunded)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if s.payment != nil {
		if err := s.payment.RefundPayment(ctx, order.ID, order.Total); err != nil {
			s.log.Error("refund processing failed", zap.String("order_id", order.ID.String()), zap.Error(err))
			return errors.Wrap(err, "refund processing failed")
		}
	}

	if err = s.repo.UpdateOrderStatusTx(ctx, tx, id, model.StatusRefunded, version); err != nil {
		if strings.Contains(err.Error(), "version conflict") {
			return ErrVersionConflict
		}
		return errors.Wrap(err, "failed to refund order")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	s.notifyAndAuditRefund(ctx, order, reason)
	metrics.OrderRefunds.Inc()
	return nil
}

// GetOrderItems retrieves items for a specific order.
func (s *service) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItemResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetOrderItems")
	defer span.End()

	_, items, _, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFoundWithID(orderID)
		}
		return nil, errors.Wrap(err, "failed to get order items")
	}

	responses := make([]model.OrderItemResponse, len(items))
	for i, item := range items {
		responses[i] = model.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			LineTotal: item.LineTotal,
		}
	}
	return responses, nil
}

// ProcessOrder moves an order to processing status.
func (s *service) ProcessOrder(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.tracer.Start(ctx, "service.ProcessOrder")
	defer span.End()

	order, _, _, err := s.repo.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFoundWithID(id)
		}
		return errors.Wrap(err, "failed to get order")
	}

	if order.Status != model.StatusPending {
		return ErrInvalidOrderStatus
	}

	return s.UpdateOrderStatus(ctx, id, model.StatusProcessing, order.Version)
}

// GetOrdersByCustomer retrieves orders for a specific customer.
func (s *service) GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]model.OrderResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetOrdersByCustomer")
	defer span.End()

	orders, err := s.repo.GetOrdersByCustomer(ctx, customerID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get orders by customer")
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		_, items, addresses, err := s.repo.GetOrder(ctx, order.ID)
		if err != nil {
			s.log.Error("failed to get order items", zap.String("order_id", order.ID.String()), zap.Error(err))
			items = []model.OrderItem{}
			addresses = []model.OrderAddress{}
		}
		responses[i] = *s.orderToResponse(&order, items, addresses)
	}
	return responses, nil
}

// GetOrdersByStatus retrieves orders by status.
func (s *service) GetOrdersByStatus(ctx context.Context, status string, limit, offset int) ([]model.OrderResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetOrdersByStatus")
	defer span.End()

	orders, err := s.repo.GetOrdersByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get orders by status")
	}

	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		_, items, addresses, err := s.repo.GetOrder(ctx, order.ID)
		if err != nil {
			s.log.Error("failed to get order items", zap.String("order_id", order.ID.String()), zap.Error(err))
			items = []model.OrderItem{}
			addresses = []model.OrderAddress{}
		}
		responses[i] = *s.orderToResponse(&order, items, addresses)
	}
	return responses, nil
}

// GetOrderStatistics retrieves order statistics for a time range.
func (s *service) GetOrderStatistics(ctx context.Context, start, end time.Time) (*model.OrderStatistics, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetOrderStatistics")
	defer span.End()

	stats, err := s.repo.GetOrderStatistics(ctx, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order statistics")
	}
	return stats, nil
}

// BulkUpdateOrderStatus updates the status of multiple orders.
func (s *service) BulkUpdateOrderStatus(ctx context.Context, orderIDs []uuid.UUID, status string) error {
	ctx, span := s.tracer.Start(ctx, "service.BulkUpdateOrderStatus")
	defer span.End()

	if len(orderIDs) == 0 {
		return nil
	}

	for _, orderID := range orderIDs {
		order, _, _, err := s.repo.GetOrder(ctx, orderID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return errors.Wrap(err, fmt.Sprintf("failed to validate order %s", orderID))
		}
		if !order.CanTransitionTo(status) {
			s.log.Warn("invalid status transition for bulk update",
				zap.String("order_id", orderID.String()),
				zap.String("current_status", order.Status),
				zap.String("target_status", status))
		}
	}

	if err := s.repo.BulkUpdateOrderStatus(ctx, orderIDs, status); err != nil {
		return errors.Wrap(err, "failed to bulk update order status")
	}

	s.auditBulkUpdate(ctx, orderIDs, status)
	metrics.OrderBulkUpdates.WithLabelValues(status).Inc()
	return nil
}

// Helper methods

// validateOrderData validates the order creation request.
func (s *service) validateOrderData(ctx context.Context, req *model.CreateOrderRequest) error {
	if len(req.Items) == 0 {
		return ErrEmptyOrderItems
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	if req.Warehouse == "" {
		return ErrInvalidWarehouse
	}
	for i, item := range req.Items {
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.UnitPrice.LessThanOrEqual(decimal.NewFromInt(0)) {
			return ErrInvalidPrice
		}
		if item.ProductID == nil && (item.SKU == nil || *item.SKU == "") {
			return fmt.Errorf("item %d must have either product_id or sku", i)
		}
	}
	return nil
}

// processOrderItems converts and validates order items.
func (s *service) processOrderItems(ctx context.Context, items []model.CreateOrderItem) ([]model.OrderItem, decimal.Decimal, error) {
	orderItems := make([]model.OrderItem, len(items))
	subtotal := decimal.NewFromInt(0)
	for i, item := range items {
		lineTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		orderItems[i] = model.OrderItem{
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			LineTotal: lineTotal,
			Version:   1,
		}
		subtotal = subtotal.Add(lineTotal)
	}
	return orderItems, subtotal, nil
}

// calculateTax calculates the tax for the order.
func (s *service) calculateTax(ctx context.Context, subtotal decimal.Decimal, customerID *uuid.UUID) (decimal.Decimal, error) {
	if s.taxCalc != nil {
		return s.taxCalc.Calculate(ctx, subtotal, customerID)
	}
	return subtotal.Mul(decimal.NewFromFloat(0.08)), nil
}

// calculateShipping calculates the shipping cost for the order.
func (s *service) calculateShipping(ctx context.Context, items []model.CreateOrderItem, customerID *uuid.UUID) (decimal.Decimal, error) {
	if s.shippingCalc != nil {
		return s.shippingCalc.Calculate(ctx, items, customerID)
	}
	return decimal.NewFromFloat(9.99), nil
}

// notifyAndAuditOrderCreated sends notifications and logs audit events for order creation.
func (s *service) notifyAndAuditOrderCreated(ctx context.Context, order *model.Order, items []model.OrderItem, customerID *uuid.UUID) {
	go func() {
		if s.notification != nil {
			if err := s.notification.SendOrderConfirmation(context.Background(), order); err != nil {
				s.log.Error("failed to send order confirmation", zap.String("order_id", order.ID.String()), zap.Error(err))
			}
		}
		if s.audit != nil {
			auditData := map[string]interface{}{
				"customer_id": customerID,
				"total":       order.Total,
				"currency":    order.Currency,
				"item_count":  len(items),
				"warehouse":   order.Warehouse,
			}
			if err := s.audit.LogOrderEvent(context.Background(), order.ID, "order_created", auditData); err != nil {
				s.log.Error("failed to log audit event", zap.Error(err))
			}
		}
	}()
}

// notifyAndAuditStatusUpdate sends notifications and logs audit events for status updates.
func (s *service) notifyAndAuditStatusUpdate(ctx context.Context, orderID uuid.UUID, newStatus, previousStatus string, version int) {
	go func() {
		if s.notification != nil {
			order := &model.Order{ID: orderID, Status: newStatus}
			if err := s.notification.SendOrderStatusUpdate(context.Background(), order, previousStatus); err != nil {
				s.log.Error("failed to send status update notification", zap.String("order_id", orderID.String()), zap.Error(err))
			}
		}
		if s.audit != nil {
			auditData := map[string]interface{}{
				"previous_status": previousStatus,
				"new_status":      newStatus,
				"version":         version,
			}
			if err := s.audit.LogOrderEvent(context.Background(), orderID, "status_updated", auditData); err != nil {
				s.log.Error("failed to log audit event", zap.Error(err))
			}
		}
	}()
}

// notifyAndAuditCancellation sends notifications and logs audit events for cancellations.
func (s *service) notifyAndAuditCancellation(ctx context.Context, order *model.Order, reason string, version int) {
	go func() {
		if s.notification != nil {
			if err := s.notification.SendOrderCancellation(context.Background(), order, reason); err != nil {
				s.log.Error("failed to send cancellation notification", zap.String("order_id", order.ID.String()), zap.Error(err))
			}
		}
		if s.audit != nil {
			auditData := map[string]interface{}{
				"reason":          reason,
				"previous_status": order.Status,
				"version":         version,
			}
			if err := s.audit.LogOrderEvent(context.Background(), order.ID, "order_cancelled", auditData); err != nil {
				s.log.Error("failed to log audit event", zap.Error(err))
			}
		}
		s.log.Info("order cancelled",
			zap.String("order_id", order.ID.String()),
			zap.String("reason", reason),
			zap.String("previous_status", order.Status))
	}()
}

// notifyAndAuditRefund logs audit events for refunds.
func (s *service) notifyAndAuditRefund(ctx context.Context, order *model.Order, reason string) {
	go func() {
		if s.audit != nil {
			auditData := map[string]interface{}{
				"reason":          reason,
				"refund_amount":   order.Total,
				"previous_status": order.Status,
			}
			if err := s.audit.LogOrderEvent(context.Background(), order.ID, "order_refunded", auditData); err != nil {
				s.log.Error("failed to log audit event", zap.Error(err))
			}
		}
		s.log.Info("order refunded",
			zap.String("order_id", order.ID.String()),
			zap.String("reason", reason),
			zap.String("amount", order.Total.String()))
	}()
}

// orderToResponse converts an order to a response DTO.
func (s *service) orderToResponse(order *model.Order, items []model.OrderItem, addresses []model.OrderAddress) *model.OrderResponse {
	itemResponses := make([]model.OrderItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = model.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			LineTotal: item.LineTotal,
		}
	}
	return &model.OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		Subtotal:   order.Subtotal,
		Tax:        order.Tax,
		Shipping:   order.Shipping,
		Total:      order.Total,
		Currency:   order.Currency,
		Items:      itemResponses,
		Addresses:  addresses,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		Version:    order.Version,
		Metadata:   order.Metadata,
	}
}