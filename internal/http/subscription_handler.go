package http

import (
	"encoding/json"
	"errors"
	"log"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/model"
	"github.com/en7ka/Effective_Mobile_testovoe.git/internal/service"
	"github.com/google/uuid"
)

const monthLayout = "01-2006"

type SubscriptionHandler struct {
	service *service.SubscriptionService
	logger  *log.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, logger *log.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) HandleSubscriptions(w nethttp.ResponseWriter, r *nethttp.Request) {
	h.logger.Printf("request method=%s path=%s", r.Method, r.URL.Path)

	switch r.Method {
	case nethttp.MethodPost:
		h.createSubscription(w, r)
	case nethttp.MethodGet:
		h.listSubscriptions(w, r)
	default:
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *SubscriptionHandler) HandleSubscriptionByID(w nethttp.ResponseWriter, r *nethttp.Request) {
	h.logger.Printf("request method=%s path=%s", r.Method, r.URL.Path)

	id, err := uuid.Parse(strings.TrimPrefix(r.URL.Path, "/subscriptions/"))
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid subscription id")
		return
	}

	switch r.Method {
	case nethttp.MethodGet:
		h.getSubscription(w, r, id)
	case nethttp.MethodPut:
		h.updateSubscription(w, r, id)
	case nethttp.MethodDelete:
		h.deleteSubscription(w, r, id)
	default:
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *SubscriptionHandler) HandleTotal(w nethttp.ResponseWriter, r *nethttp.Request) {
	h.logger.Printf("request method=%s path=%s", r.Method, r.URL.Path)

	if r.Method != nethttp.MethodGet {
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()

	startDate, err := parseMonth(query.Get("start_date"))
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid start_date, use MM-YYYY")
		return
	}

	endDate, err := parseMonth(query.Get("end_date"))
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid end_date, use MM-YYYY")
		return
	}

	filter, ok := h.filterFromQuery(w, r)
	if !ok {
		return
	}
	filter.PeriodStart = &startDate
	filter.PeriodEnd = &endDate

	total, err := h.service.Total(r.Context(), filter)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]int{"total": total})
}

func (h *SubscriptionHandler) createSubscription(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request subscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid json")
		return
	}

	subscription, err := request.toModel(uuid.Nil)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	subscription, err = h.service.Create(r.Context(), subscription)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("created subscription id=%s user_id=%s service=%s", subscription.ID, subscription.UserID, subscription.ServiceName)
	writeJSON(w, nethttp.StatusCreated, toResponse(subscription))
}

func (h *SubscriptionHandler) getSubscription(w nethttp.ResponseWriter, r *nethttp.Request, id uuid.UUID) {
	subscription, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, toResponse(subscription))
}

func (h *SubscriptionHandler) listSubscriptions(w nethttp.ResponseWriter, r *nethttp.Request) {
	filter, ok := h.filterFromQuery(w, r)
	if !ok {
		return
	}

	subscriptions, err := h.service.List(r.Context(), filter)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response := make([]subscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		response = append(response, toResponse(subscription))
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func (h *SubscriptionHandler) updateSubscription(w nethttp.ResponseWriter, r *nethttp.Request, id uuid.UUID) {
	var request subscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid json")
		return
	}

	subscription, err := request.toModel(id)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	subscription, err = h.service.Update(r.Context(), subscription)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("updated subscription id=%s", subscription.ID)
	writeJSON(w, nethttp.StatusOK, toResponse(subscription))
}

func (h *SubscriptionHandler) deleteSubscription(w nethttp.ResponseWriter, r *nethttp.Request, id uuid.UUID) {
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("deleted subscription id=%s", id)
	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *SubscriptionHandler) filterFromQuery(w nethttp.ResponseWriter, r *nethttp.Request) (model.SubscriptionFilter, bool) {
	var filter model.SubscriptionFilter
	query := r.URL.Query()

	if userID := query.Get("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			writeError(w, nethttp.StatusBadRequest, "invalid user_id")
			return model.SubscriptionFilter{}, false
		}

		filter.UserID = &id
	}

	filter.ServiceName = query.Get("service_name")
	return filter, true
}

func (h *SubscriptionHandler) handleServiceError(w nethttp.ResponseWriter, err error) {
	h.logger.Printf("request error: %v", err)

	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, nethttp.StatusNotFound, "subscription not found")
	case errors.Is(err, service.ErrValidation):
		writeError(w, nethttp.StatusBadRequest, err.Error())
	default:
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
	}
}

type subscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (r subscriptionRequest) toModel(id uuid.UUID) (model.Subscription, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return model.Subscription{}, errors.New("invalid user_id")
	}

	startDate, err := parseMonth(r.StartDate)
	if err != nil {
		return model.Subscription{}, errors.New("invalid start_date, use MM-YYYY")
	}

	var endDate *time.Time
	if r.EndDate != nil && *r.EndDate != "" {
		parsedEndDate, err := parseMonth(*r.EndDate)
		if err != nil {
			return model.Subscription{}, errors.New("invalid end_date, use MM-YYYY")
		}

		endDate = &parsedEndDate
	}

	return model.Subscription{
		ID:          id,
		ServiceName: strings.TrimSpace(r.ServiceName),
		Price:       r.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type subscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toResponse(subscription model.Subscription) subscriptionResponse {
	var endDate *string
	if subscription.EndDate != nil {
		value := formatMonth(*subscription.EndDate)
		endDate = &value
	}

	return subscriptionResponse{
		ID:          subscription.ID.String(),
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID.String(),
		StartDate:   formatMonth(subscription.StartDate),
		EndDate:     endDate,
		CreatedAt:   subscription.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   subscription.UpdatedAt.Format(time.RFC3339),
	}
}

func parseMonth(value string) (time.Time, error) {
	return time.Parse(monthLayout, value)
}

func formatMonth(value time.Time) string {
	return value.Format(monthLayout)
}

func writeJSON(w nethttp.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w nethttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
