package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayment(t *testing.T) {
	t.Run("Get payment not found", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/payments/1", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("List payments not found", func(t *testing.T) {
		clearDB(t)

		token, err := registerAndLogin(t,
			"tester",
			"password123",
			"customer",
		)
		assert.NoError(t, err)

		w, err := doRequest("GET", "/payments", nil, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
