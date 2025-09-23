package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCheckError(t *testing.T) {
	// Test with a nil error
	CheckError(nil) // Should not panic

	// Test with a non-nil error
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("CheckError did not panic on non-nil error")
		}
	}()
	CheckError(fmt.Errorf("test error"))
}

func TestRespondJSON(t *testing.T) {
	// Create a ResponseRecorder to capture the response
	recorder := httptest.NewRecorder()

	// Call RespondJSON
	RespondJSON(recorder, http.StatusOK, "success", "", map[string]string{"message": "success"})

	// Check the response status code
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("RespondJSON returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"status":"success","message":"","data":{"message":"success"}}
`
	actual := recorder.Body.String()
	if actual != expected {
		t.Errorf("RespondJSON returned wrong body:\nGot: %q\nWant: %q", actual, expected)
	}
}

func TestRespondError(t *testing.T) {
	// Create a ResponseRecorder to capture the response
	recorder := httptest.NewRecorder()

	// Call RespondError
	RespondError(recorder, http.StatusInternalServerError, "error occurred")

	// Check the response status code
	if status := recorder.Code; status != http.StatusInternalServerError {
		t.Errorf("RespondError returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expected := `{"status":"error","message":"error occurred"}
`
	actual := recorder.Body.String()
	if actual != expected {
		t.Errorf("RespondError returned wrong body:\nGot: %q\nWant: %q", actual, expected)
	}
}

func TestResponseMessage(t *testing.T) {
	// Call ResponseMessage
	result := ResponseMessage("Test message")

	// Verify the result
	expected := map[string]string{"message": "Test message"}
	if result["message"] != expected["message"] {
		t.Errorf("ResponseMessage returned wrong result: got %v want %v", result, expected)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	// Create a new gin context for testing
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test when no user ID is set
	id := GetUserIDFromContext(c)
	if id != "" {
		t.Errorf("Expected empty user ID, got %s", id)
	}

	// Test when user ID is set correctly
	expectedID := "test-user-id"
	c.Set("user_id", expectedID)
	id = GetUserIDFromContext(c)
	if id != expectedID {
		t.Errorf("Expected user ID %s, got %s", expectedID, id)
	}

	// Test when user ID is set but with wrong type
	c.Set("user_id", 123) // not a string
	id = GetUserIDFromContext(c)
	if id != "" {
		t.Errorf("Expected empty user ID when type is wrong, got %s", id)
	}
}

func TestIsAdminUser(t *testing.T) {
	// Create a new gin context for testing
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test when admin flag is not set
	isAdmin := IsAdminUser(c)
	if isAdmin {
		t.Errorf("Expected false for admin check, got true")
	}

	// Test when admin flag is set to true
	c.Set("is_admin", true)
	isAdmin = IsAdminUser(c)
	if !isAdmin {
		t.Errorf("Expected true for admin check, got false")
	}

	// Test when admin flag is set to false
	c.Set("is_admin", false)
	isAdmin = IsAdminUser(c)
	if isAdmin {
		t.Errorf("Expected false for admin check, got true")
	}

	// Test when admin flag is set but with wrong type
	c.Set("is_admin", "true") // not a bool
	isAdmin = IsAdminUser(c)
	if isAdmin {
		t.Errorf("Expected false for admin check when type is wrong, got true")
	}
}
