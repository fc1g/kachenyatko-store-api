package order

import (
	"context"
	"net/http"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/httpx"
	"github.com/google/uuid"
)

const (
	orderIDParam          = "orderID"
	toyIDParam            = "toyID"
	orderStatusQueryParam = "status"
)

type service interface {
	Create(ctx context.Context, toyID *uuid.UUID, input createOrderInput) (sqlc.Order, error)
	List(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error)
	ListCustom(ctx context.Context, filter orderFilter, limit, offset int32) (orderPage, error)
	ListByToy(ctx context.Context, toyID uuid.UUID, filter orderFilter, limit, offset int32) (orderPage, error)
	Get(ctx context.Context, id uuid.UUID) (sqlc.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, input updateOrderInput) (sqlc.Order, error)
}

type Handler struct {
	svc service
}

func NewHandler(svc service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// CreateCustom godoc
//
// @Summary      Place a custom order
// @Description  Order a toy that is not in the catalog. The message describing the toy is required. Limited to 10 orders per 10 minutes per client IP across both order endpoints; exceeding it returns 429
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      createOrderInput  true  "Contact details and description of the toy"
// @Success      201    {object}  orderResponse
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      413    {object}  httpx.ErrorResponse
// @Failure      429    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /orders [post]
func (h *Handler) CreateCustom(w http.ResponseWriter, r *http.Request) {
	h.create(w, r, nil)
}

// CreateForToy godoc
//
// @Summary      Order a toy
// @Description  Place an order for a toy from the catalog. The message is optional. Limited to 10 orders per 10 minutes per client IP across both order endpoints; exceeding it returns 429
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        toyID  path      string                    true  "Toy ID"  format(uuid)
// @Param        order  body      createOrderInput  true  "Contact details and optional message"
// @Success      201    {object}  orderResponse
// @Failure      400    {object}  httpx.ErrorResponse
// @Failure      404    {object}  httpx.ErrorResponse
// @Failure      413    {object}  httpx.ErrorResponse
// @Failure      429    {object}  httpx.ErrorResponse
// @Failure      500    {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/orders [post]
func (h *Handler) CreateForToy(w http.ResponseWriter, r *http.Request) {
	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	h.create(w, r, &toyID)
}

// List godoc
//
// @Summary      List orders
// @Description  All orders, both for catalog toys and custom ones, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {object}  orderPageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /orders [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := httpx.ParseLimit(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	offset, err := httpx.ParseOffset(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var orderStatus *string
	if val := r.URL.Query().Get(orderStatusQueryParam); val != "" {
		orderStatus = &val
	}

	page, err := h.svc.List(r.Context(), orderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderPageResponse(page))
}

// ListCustomOrders godoc
//
// @Summary      List custom orders
// @Description  Orders for toys that are not in the catalog, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {object}  orderPageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /orders/custom [get]
func (h *Handler) ListCustom(w http.ResponseWriter, r *http.Request) {
	limit, err := httpx.ParseLimit(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	offset, err := httpx.ParseOffset(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var orderStatus *string = nil
	if val := r.URL.Query().Get(orderStatusQueryParam); val != "" {
		orderStatus = &val
	}

	page, err := h.svc.ListCustom(r.Context(), orderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderPageResponse(page))
}

// ListOrdersByToyID godoc
//
// @Summary      List orders for a toy
// @Description  Orders for a catalog toy, sorted from newest to oldest, paginated with limit/offset
// @Tags         orders
// @Produce      json
// @Param        toyID   path      string  true   "Toy ID"                   format(uuid)
// @Param        status  query     string  false  "Filter by order status"   Enums(open, closed)
// @Param        limit   query     int     false  "Page size"                default(9)  minimum(1) maximum(100)
// @Param        offset  query     int     false  "Number of items to skip"  default(0)  minimum(0)
// @Success      200     {object}  orderPageResponse
// @Failure      400     {object}  httpx.ErrorResponse
// @Failure      500     {object}  httpx.ErrorResponse
// @Router       /toys/{toyID}/orders [get]
func (h *Handler) ListByToy(w http.ResponseWriter, r *http.Request) {
	toyID, err := httpx.ParseIDParam(r, toyIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	limit, err := httpx.ParseLimit(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	offset, err := httpx.ParseOffset(r)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var orderStatus *string = nil
	if val := r.URL.Query().Get(orderStatusQueryParam); val != "" {
		orderStatus = &val
	}

	page, err := h.svc.ListByToy(r.Context(), toyID, orderFilter{
		Status: orderStatus,
	}, limit, offset)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderPageResponse(page))
}

// Get godoc
//
// @Summary      Get order
// @Description  Get an order by its ID, either for a catalog toy or a custom one
// @Tags         orders
// @Produce      json
// @Param        orderID  path      string  true  "Order ID"  format(uuid)
// @Success      200      {object}  orderResponse
// @Failure      400      {object}  httpx.ErrorResponse
// @Failure      404      {object}  httpx.ErrorResponse
// @Failure      500      {object}  httpx.ErrorResponse
// @Router       /orders/{orderID} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, orderIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	order, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderResponse(order))
}

// UpdateStatus godoc
//
// @Summary      Update order status
// @Description  Change the status of an order by its ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        orderID  path      string                    true  "Order ID"  format(uuid)
// @Param        status   body      updateOrderInput  true  "New status"
// @Success      200      {object}  orderResponse
// @Failure      400      {object}  httpx.ErrorResponse
// @Failure      404      {object}  httpx.ErrorResponse
// @Failure      413      {object}  httpx.ErrorResponse
// @Failure      500      {object}  httpx.ErrorResponse
// @Router       /orders/{orderID}/status [patch]
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIDParam(r, orderIDParam)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	var input updateOrderInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	order, err := h.svc.UpdateStatus(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toOrderResponse(order))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, toyID *uuid.UUID) {
	var input createOrderInput
	if err := httpx.DecodeJSON(r.Body, &input); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	order, err := h.svc.Create(r.Context(), toyID, input)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toOrderResponse(order))
}
