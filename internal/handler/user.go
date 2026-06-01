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

type UserHandler struct {
	service            *service.UserService
	validate           *validator.Validate
	maxPaginationLimit int
}

func NewUserHandler(service *service.UserService, validate *validator.Validate, maxPaginationLimit int) *UserHandler {
	return &UserHandler{
		service:            service,
		validate:           validate,
		maxPaginationLimit: maxPaginationLimit,
	}
}

func (h *UserHandler) getUserID(w http.ResponseWriter, r *http.Request) (int, bool) {
	userIDString := chi.URLParam(r, "user_id")

	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		SendError(w, http.StatusBadRequest, "некорректный идентификатор пользователя")
		return 0, false
	}

	return userID, true
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	user, err := h.service.Get(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			SendError(w, http.StatusNotFound, "данного пользователя не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitString)
	if err != nil || limit <= 0 || limit > h.maxPaginationLimit {
		limit = h.maxPaginationLimit
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) Block(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	user, err := h.service.Block(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			SendError(w, http.StatusNotFound, "данного пользователя не существует")
			return
		} else if errors.Is(err, domain.ErrUserInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "пользователь уже заблокирован")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	user, err := h.service.Unblock(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			SendError(w, http.StatusNotFound, "данного пользователя не существует")
			return
		} else if errors.Is(err, domain.ErrUserInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "пользователь уже разблокирован")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) SetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	type setBalanceRequest struct {
		Amount int `json:"amount" default:"0" validate:"gte=0"`
	}

	var request setBalanceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		SendError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}

	err = h.validate.Struct(request)
	if err != nil {
		SendError(w, http.StatusUnprocessableEntity, "данные не прошли валидацию")
		return
	}

	user, err := h.service.SetBalance(r.Context(), userID, request.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			SendError(w, http.StatusNotFound, "данного пользователя не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
