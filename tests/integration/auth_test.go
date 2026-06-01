package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	t.Run("Register success", func(t *testing.T) {
		clearDB(t)

		body := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "customer",
		}
		w, err := doRequest("POST", "/auth/register", body, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Register duplicate", func(t *testing.T) {
		clearDB(t)

		bodyOne := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "customer",
		}
		w, err := doRequest("POST", "/auth/register", bodyOne, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bodyTwo := map[string]any{
			"username": "tester",
			"password": "password765",
			"role":     "performer",
		}
		w, err = doRequest("POST", "/auth/register", bodyTwo, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Register short name", func(t *testing.T) {
		clearDB(t)

		bodyOne := map[string]any{
			"username": "test",
			"password": "password123",
			"role":     "customer",
		}
		w, err := doRequest("POST", "/auth/register", bodyOne, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Register incorrect role", func(t *testing.T) {
		clearDB(t)

		bodyOne := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "transformer",
		}
		w, err := doRequest("POST", "/auth/register", bodyOne, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Register admin", func(t *testing.T) {
		clearDB(t)

		bodyOne := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "admin",
		}
		w, err := doRequest("POST", "/auth/register", bodyOne, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Login success", func(t *testing.T) {
		clearDB(t)

		bodyRegister := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "customer",
		}
		w, err := doRequest("POST", "/auth/register", bodyRegister, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bodyLogin := map[string]any{
			"username": "tester",
			"password": "password123",
		}
		w, err = doRequest("POST", "/auth/login", bodyLogin, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Login wrong password", func(t *testing.T) {
		clearDB(t)

		bodyRegister := map[string]any{
			"username": "tester",
			"password": "password123",
			"role":     "customer",
		}
		w, err := doRequest("POST", "/auth/register", bodyRegister, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		bodyLogin := map[string]any{
			"username": "tester",
			"password": "password321",
		}
		w, err = doRequest("POST", "/auth/login", bodyLogin, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Login doesn't exist", func(t *testing.T) {
		clearDB(t)

		bodyLogin := map[string]any{
			"username": "tester",
			"password": "password123",
		}
		w, err := doRequest("POST", "/auth/login", bodyLogin, "")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
