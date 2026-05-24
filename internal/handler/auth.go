package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"ssubench/internal/domain"
	"ssubench/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	service  *service.AuthService
	validate *validator.Validate
}

func NewAuthHandler(service *service.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		service:  service,
		validate: validate,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	type registerRequest struct {
		Username string          `json:"username" validate:"required,min=5,max=16"`
		Password string          `json:"password" validate:"required,min=8,max=32"`
		Role     domain.UserRole `json:"role" validate:"required,oneof=customer performer"`
	}

	var request registerRequest

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

	user, err := h.service.Register(r.Context(), request.Username, request.Password, request.Role)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			SendError(w, http.StatusConflict, "данное имя пользователя уже занято")
			return
		}

		log.Printf("unhandled error: %v", fmt.Errorf("AuthHandler.Register: %w", err))
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Username string `json:"username" validate:"required,min=5,max=16"`
		Password string `json:"password" validate:"required,min=8,max=32"`
	}

	var request loginRequest

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

	token, err := h.service.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			SendError(w, http.StatusUnauthorized, "некорректные данные")
		}

		log.Printf("unhandled error: %v", fmt.Errorf("AuthHandler.Login: %w", err))
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) DumbCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("All good")
}
