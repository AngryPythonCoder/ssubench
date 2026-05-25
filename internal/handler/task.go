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

type TaskHandler struct {
	service            *service.TaskService
	validate           *validator.Validate
	maxPaginationLimit int
}

func NewTaskHandler(service *service.TaskService, validate *validator.Validate, maxPaginationLimit int) *TaskHandler {
	return &TaskHandler{
		service:            service,
		validate:           validate,
		maxPaginationLimit: maxPaginationLimit,
	}
}

func (h *TaskHandler) getTaskID(w http.ResponseWriter, r *http.Request) (int, bool) {
	taskIDString := chi.URLParam(r, "task_id")

	taskID, err := strconv.Atoi(taskIDString)
	if err != nil {
		SendError(w, http.StatusBadRequest, "некорректный идентификатор задачи")
		return 0, false
	}

	return taskID, true
}

func (h *TaskHandler) getBidID(w http.ResponseWriter, r *http.Request) (int, bool) {
	bidIDString := chi.URLParam(r, "bid_id")

	bidID, err := strconv.Atoi(bidIDString)
	if err != nil {
		SendError(w, http.StatusBadRequest, "некорректный идентификатор отклика")
		return 0, false
	}

	return bidID, true
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(domain.UserIDKey).(int)

	type createTaskRequest struct {
		Title       string `json:"title" validate:"required,min=5,max=128"`
		Description string `json:"description" validate:"required,max=65536"`
		Reward      int    `json:"reward" validate:"required,gt=0"`
	}

	var request createTaskRequest

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

	task, err := h.service.Create(
		r.Context(),
		request.Title,
		request.Description,
		request.Reward,
		userID,
	)

	if err != nil {
		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	task, err := h.service.Get(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitString)
	if err != nil || limit <= 0 {
		limit = h.maxPaginationLimit
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil || offset < 0 {
		offset = 0
	}

	tasks, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) CreateBid(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	userID := r.Context().Value(domain.UserIDKey).(int)

	type createBidRequest struct {
		Text string `json:"text" validate:"required,max=4096"`
	}

	var request createBidRequest

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

	bid, err := h.service.CreateBid(
		r.Context(),
		taskID,
		userID,
		request.Text,
	)

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrBidAlreadyExists) {
			SendError(w, http.StatusConflict, "вы уже откликнулись на данную задачу")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bid)
}

func (h *TaskHandler) GetBid(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	bidID, ok := h.getBidID(w, r)
	if !ok {
		return
	}

	bid, err := h.service.GetBid(r.Context(), taskID, bidID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrBidNotFound) {
			SendError(w, http.StatusNotFound, "данного отклика не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid)
}

func (h *TaskHandler) ListBids(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitString)
	if err != nil || limit <= 0 {
		limit = h.maxPaginationLimit
	}

	offset, err := strconv.Atoi(offsetString)
	if err != nil || offset < 0 {
		offset = 0
	}

	bids, err := h.service.ListBids(r.Context(), taskID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)
}

func (h *TaskHandler) AcceptBid(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	bidID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	userID := r.Context().Value(domain.UserIDKey).(int)

	task, err := h.service.AcceptBid(r.Context(), taskID, userID, bidID)

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrBidNotFound) {
			SendError(w, http.StatusForbidden, "данного отклика не существует")
			return
		} else if errors.Is(err, domain.ErrNotAuthorized) {
			SendError(w, http.StatusForbidden, "нет прав, чтобы принять отклик")
			return
		} else if errors.Is(err, domain.ErrTaskInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "невозможно принять отклик")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Finish(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	userID := r.Context().Value(domain.UserIDKey).(int)

	task, err := h.service.Finish(r.Context(), taskID, userID)

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrNotAuthorized) {
			SendError(w, http.StatusForbidden, "нет прав, чтобы завершить задачу")
			return
		} else if errors.Is(err, domain.ErrTaskInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "невозможно завершить задачу")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	userID := r.Context().Value(domain.UserIDKey).(int)

	task, err := h.service.Complete(r.Context(), taskID, userID)

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrNotAuthorized) {
			SendError(w, http.StatusForbidden, "нет прав, чтобы подтвердить выполнение задачи")
			return
		} else if errors.Is(err, domain.ErrTaskInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "невозможно подтвердить выполнение задачи")
			return
		} else if errors.Is(err, domain.ErrInsufficientFunds) {
			SendError(w, http.StatusConflict, "недостаточно средств для подтверждения выполнения задачи")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	taskID, ok := h.getTaskID(w, r)
	if !ok {
		return
	}

	userID := r.Context().Value(domain.UserIDKey).(int)

	task, err := h.service.Cancel(r.Context(), taskID, userID)

	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			SendError(w, http.StatusNotFound, "данной задачи не существует")
			return
		} else if errors.Is(err, domain.ErrNotAuthorized) {
			SendError(w, http.StatusForbidden, "нет прав, чтобы отменить задачу")
			return
		} else if errors.Is(err, domain.ErrTaskInvalidStatusTransition) {
			SendError(w, http.StatusConflict, "невозможно отменить задачу")
			return
		}

		log.Printf("unhandled error: %v", err)
		SendError(w, http.StatusInternalServerError, "что-то пошло не так")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}
