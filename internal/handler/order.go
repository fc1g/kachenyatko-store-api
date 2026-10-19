package handler

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
	"github.com/fc1g/kachenyatko-store-api/internal/service"
	"github.com/google/uuid"
)

type orderService interface {
	Create(ctx context.Context, toyID *uuid.UUID, input service.CreateOrderInput) (repository.Order, error)
	GetAll(ctx context.Context, filter service.OrderFilter, limit, offset int32) ([]repository.Order, error)
	GetAllCustomOrders(ctx context.Context, filter service.OrderFilter, limit, offset int32) ([]repository.Order, error)
	GetAllOrdersByToyID(ctx context.Context, toyID uuid.UUID, filter service.OrderFilter, limit, offset int32) ([]repository.Order, error)
	GetOne(ctx context.Context, id uuid.UUID) (repository.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, input service.UpdateOrderInput) (repository.Order, error)
}

type OrderHandler struct {
	service orderService
}

func NewOrderHandler(svc orderService) *OrderHandler {
	return &OrderHandler{
		service: svc,
	}
}

// CreateForToy godoc
//
// @Summary      Order a toy
// @Description  Place an order for a toy from the catalog. The message is optional
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        toyID  path      string                    true  "Toy ID"  format(uuid)
// @Param        order  body      service.CreateOrderInput  true  "Contact details and optional message"
// @Success      201    {object}  orderResponse
// @Failure      400    {object}  errorResponse
// @Failure      404    {object}  errorResponse
// @Failure      413    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /toys/{toyID}/orders [post]
func (h *OrderHandler) CreateForToy(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	h.create(w, r, &toyID)
}

// CreateCustom godoc
//
// @Summary      Place a custom order
// @Description  Order a toy that is not in the catalog. The message describing the toy is required
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      service.CreateOrderInput  true  "Contact details and description of the toy"
// @Success      201    {object}  orderResponse
// @Failure      400    {object}  errorResponse
// @Failure      413    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Router       /orders [post]
func (h *OrderHandler) CreateCustom(w http.ResponseWriter, r *http.Request) {
	h.create(w, r, nil)
}

func (h *OrderHandler) create(w http.ResponseWriter, r *http.Request, toyID *uuid.UUID) {
	var input service.CreateOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}

	order, err := h.service.Create(r.Context(), toyID, input)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toOrderResponse(order))
}

// GetAll godoc
//
// @Summary      List orders
// @Description  All orders, both for catalog toys and custom ones, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   orderResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /orders [get]
func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	orderStatus := parseOptionalStringQuery(r.URL.Query(), OrderStatusQueryParam)

	orders, err := h.service.GetAll(r.Context(), service.OrderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponses(orders))
}

// GetAllCustomOrders godoc
//
// @Summary      List custom orders
// @Description  Orders for toys that are not in the catalog, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   orderResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /orders/custom [get]
func (h *OrderHandler) GetAllCustomOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	orderStatus := parseOptionalStringQuery(r.URL.Query(), OrderStatusQueryParam)

	orders, err := h.service.GetAllCustomOrders(r.Context(), service.OrderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponses(orders))
}

// GetAllOrdersByToyID godoc
//
// @Summary      List orders for a toy
// @Description  Orders for a catalog toy, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        toyID   path      string  true   "Toy ID"                   format(uuid)
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(20) minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {array}   orderResponse
// @Failure      400     {object}  errorResponse
// @Failure      500     {object}  errorResponse
// @Router       /toys/{toyID}/orders [get]
func (h *OrderHandler) GetAllOrdersByToyID(w http.ResponseWriter, r *http.Request) {
	toyID, err := parseIDParam(r, ToyIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	orderStatus := parseOptionalStringQuery(r.URL.Query(), OrderStatusQueryParam)

	orders, err := h.service.GetAllOrdersByToyID(r.Context(), toyID, service.OrderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponses(orders))
}

// GetOne godoc
//
// @Summary      Get order
// @Description  Get an order by its ID, either for a catalog toy or a custom one
// @Tags         orders
// @Produce      json
// @Param        orderID  path      string  true  "Order ID"  format(uuid)
// @Success      200      {object}  orderResponse
// @Failure      400      {object}  errorResponse
// @Failure      404      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /orders/{orderID} [get]
func (h *OrderHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, OrderIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	order, err := h.service.GetOne(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponse(order))
}

// UpdateStatus godoc
//
// @Summary      Update order status
// @Description  Change the status of an order by its ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        orderID  path      string                    true  "Order ID"  format(uuid)
// @Param        status   body      service.UpdateOrderInput  true  "New status"
// @Success      200      {object}  orderResponse
// @Failure      400      {object}  errorResponse
// @Failure      404      {object}  errorResponse
// @Failure      413      {object}  errorResponse
// @Failure      500      {object}  errorResponse
// @Router       /orders/{orderID}/status [patch]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, OrderIDParam)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var input service.UpdateOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}

	order, err := h.service.UpdateStatus(r.Context(), id, input)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponse(order))
}
