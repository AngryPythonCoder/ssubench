package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/service"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type PaymentHandler struct {
	service            *service.PaymentService
	validate           *validator.Validate
	maxPaginationLimit int
}

func NewPaymentHandler(service *service.PaymentService, validate *validator.Validate, maxPaginationLimit int) *PaymentHandler {
	return &PaymentHandler{
		service:            service,
		validate:           validate,
		maxPaginationLimit: maxPaginationLimit,
	}
}

func (h *PaymentHandler) getPaymentID(w http.ResponseWriter, r *http.Request) (int, bool) {
	paymentIDString := chi.URLParam(r, "payment_id")

	paymentID, err := strconv.Atoi(paymentIDString)
	if err != nil {
		SendError(w, http.StatusBadRequest, "некорректный идентификатор оплаты")
		return 0, false
	}

	return paymentID, true
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := h.getPaymentID(w, r)
	if !ok {
		return
	}

	payment, err := h.service.Get(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			SendError(w, http.StatusNotFound, "данной оплаты не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = h.maxPaginationLimit
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil || offset < 0 {
		offset = 0
	}

	payments, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(payments)
}
