package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockContext struct {
	username string
	password string
}

func (m *MockContext) FormValue(name string) string {
	switch name {
	case "username":
		return m.username
	case "password":
		return m.password
	default:
		return ""
	}
}

func TestAuthenticate(t *testing.T) {

	oldUsername := os.Getenv("ADMIN_USERNAME")
	oldPassword := os.Getenv("ADMIN_PASSWORD")
	defer func() {
		os.Setenv("ADMIN_USERNAME", oldUsername)
		os.Setenv("ADMIN_PASSWORD", oldPassword)
	}()

	tests := []struct {
		name          string
		username      string
		password      string
		passwordTest  string
		expectedAuth  bool
		expectedError error
	}{
		{
			name:          "Valid credentials",
			username:      "admin",
			password:      "password",
			passwordTest:  "password",
			expectedAuth:  true,
			expectedError: nil,
		},
		{
			name:          "Invalid credentials",
			username:      "user",
			password:      "password",
			passwordTest:  "wrongpassword",
			expectedAuth:  false,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			os.Setenv("ADMIN_USERNAME", test.username)
			os.Setenv("ADMIN_PASSWORD", test.password)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/admin/deduction", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			auth, err := authenticate(test.username, test.passwordTest, c)

			assert.Equal(t, test.expectedAuth, auth)

			if test.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
