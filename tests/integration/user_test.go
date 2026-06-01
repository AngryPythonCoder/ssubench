package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	t.Run("Get user success", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/users/1", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get user not found", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/users/2", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List users success", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/users", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Block user success", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/block", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Block user already blocked", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/block", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/users/2/block", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Block user not admin", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"not_admin",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/block", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Unblock user success", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/block", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		w, err = doRequest("POST", "/users/2/unblock", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unblock user already unblocked", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/unblock", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Unlock user not admin", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"not_admin",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		_, err = registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/2/unblock", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Set balance success", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/1/set-balance", map[string]any{"amount": 5}, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Set balance zero", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/1/set-balance", map[string]any{"amount": 0}, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Set balance below zero", func(t *testing.T) {
		clearDB(t)

		token, err := addAdminAndLogin(t,
			"admin",
			"password123",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/1/set-balance", map[string]any{"amount": -7}, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Set balance not admin", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("POST", "/users/1/set-balance", map[string]any{"amount": 5}, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
