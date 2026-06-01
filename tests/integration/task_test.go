package integration_test

import (
	"encoding/json"
	"net/http"
	"ssubench/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	t.Run("Main flow", func(t *testing.T) {
		clearDB(t)

		adminToken, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/tasks/1/finish", nil, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/users/2/set-balance", map[string]any{"amount": 100}, adminToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/tasks/1/confirm", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("GET", "/payments/1", nil, adminToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var responseUser struct {
			ID        int
			Username  string            `json:"username"`
			Role      domain.UserRole   `json:"role"`
			Status    domain.UserStatus `json:"status"`
			Balance   int               `json:"balance"`
			CreatedAt time.Time         `json:"created_at"`
		}

		w, err = doRequest("GET", "/users/2", nil, adminToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, responseUser.Balance, 0)

		w, err = doRequest("GET", "/users/3", nil, adminToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, responseUser.Balance, 100)
	})

	t.Run("Create task success", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		body := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", body, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Create task short name", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		body := map[string]any{
			"title":       "t",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", body, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Create task not customer", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		body := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", body, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Get task success", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		body := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", body, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("GET", "/tasks/1", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get task not found", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/tasks/1", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List tasks success", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/tasks", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Create bid success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Create bid already responded", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Get bid success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("GET", "/tasks/1/bids/1", nil, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get bid not found", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("GET", "/tasks/1/bids/1", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Get bid task not found", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/tasks/1/bids/1", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List bids success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("GET", "/tasks/1/bids", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List bids task not found", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/tasks/1/bids", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Accept bid success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Accept bid already accepted", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		anotherPerformerToken, err := registerAndLogin(t,
			"performer2",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, anotherPerformerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/2/accept", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Accept bid not creator", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		anotherCustomerToken, err := registerAndLogin(t,
			"customer2",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", nil, anotherCustomerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Finish task success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/bids/1/accept", bidBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/tasks/1/finish", nil, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Finish task not assigned", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		anotherPerformerToken, err := registerAndLogin(t,
			"performer2",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/finish", nil, anotherPerformerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Cancel task success", func(t *testing.T) {
		clearDB(t)

		customerToken, err := registerAndLogin(t,
			"customer",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		performerToken, err := registerAndLogin(t,
			"performer",
			"password123",
			"performer",
		)
		assert.NoError(t, err)

		taskBody := map[string]any{
			"title":       "title",
			"description": "decription",
			"reward":      100,
		}

		w, err := doRequest("POST", "/tasks", taskBody, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bidBody := map[string]any{
			"text": "some text",
		}

		w, err = doRequest("POST", "/tasks/1/respond", bidBody, performerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		w, err = doRequest("POST", "/tasks/1/cancel", nil, customerToken)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
