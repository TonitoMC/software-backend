//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRegistrationFlow(t *testing.T) {
	fmt.Println("=== TestUserRegistrationFlow starting ===")

	// Initialize test setup
	initTest()

	registerReq := map[string]string{
		"email":    "test@example.com",
		"username": "testuser",
		"password": "password123",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	testApp.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var registerResp RegisterResponse
	err := json.Unmarshal(rec.Body.Bytes(), &registerResp)
	require.NoError(t, err)

	assert.Equal(t, "testuser", registerResp.Username)
	assert.Equal(t, "test@example.com", registerResp.Email)
	assert.Contains(t, registerResp.Roles, "user")
}
