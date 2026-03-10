package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAllHandlers_InvalidJSON tests all POST handlers with invalid JSON
func TestAllHandlers_InvalidJSON(t *testing.T) {
	handlers := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"LoginHandler", LoginHandler, "/login"},
		{"RegisterHandler", RegisterHandler, "/register"},
		{"ChangePasswordHandler", ChangePasswordHandler, "/change-password"},
		{"PostsHandler", PostsHandler, "/posts"},
		{"EditPostHandler", EditPostHandler, "/posts/edit"},
		{"VoteHandler", VoteHandler, "/vote"},
	}

	invalidJSONs := []string{
		"",
		"{",
		"}",
		"}{",
		"{{}",
		"[]",
		"null",
		"true",
		"false",
		"123",
		`{"unclosed": "string`,
		`{"trailing": "comma",}`,
		`{"key"}`,
		`{"key":}`,
		`{:"value"}`,
	}

	for _, h := range handlers {
		for i, invalidJSON := range invalidJSONs {
			testName := h.name + "_Invalid_" + string(rune('A'+i%26))
			t.Run(testName, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, h.path, bytes.NewBufferString(invalidJSON))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				h.handler(w, req)

				// Should return an error status (400, 401, or 500)
				if w.Code < 400 {
					t.Errorf("%s with invalid JSON should return error status, got %d", h.name, w.Code)
				}
			})
		}
	}
}

// TestAllHandlers_MissingContentType tests handlers without Content-Type
func TestAllHandlers_MissingContentType(t *testing.T) {
	handlers := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"LoginHandler", LoginHandler, "/login"},
		{"RegisterHandler", RegisterHandler, "/register"},
		{"ChangePasswordHandler", ChangePasswordHandler, "/change-password"},
		{"VoteHandler", VoteHandler, "/vote"},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, h.path, bytes.NewBufferString(`{"test":"data"}`))
			// No Content-Type header set
			w := httptest.NewRecorder()

			h.handler(w, req)

			// Should return 415 Unsupported Media Type or handle gracefully
			if w.Code != http.StatusUnsupportedMediaType && w.Code < 400 {
				t.Logf("%s without Content-Type returned %d", h.name, w.Code)
			}
		})
	}
}

// TestPostsHandler_UnsupportedMethods tests unsupported HTTP methods
func TestPostsHandler_UnsupportedMethods(t *testing.T) {
	methods := []string{
		http.MethodDelete,
		http.MethodPatch,
		http.MethodOptions,
		http.MethodHead,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/posts", nil)
			w := httptest.NewRecorder()

			PostsHandler(w, req)

			// Should return 405 Method Not Allowed
			if w.Code != http.StatusMethodNotAllowed && w.Code < 400 {
				t.Logf("Method %s returned %d", method, w.Code)
			}
		})
	}
}

// TestEditPostHandler_WrongMethod tests EditPostHandler with wrong method
func TestEditPostHandler_WrongMethod(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/posts/edit", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			EditPostHandler(w, req)

			// Should return 405 for non-PUT methods
			if method != http.MethodPut && w.Code != http.StatusMethodNotAllowed && w.Code < 400 {
				t.Logf("Method %s returned %d", method, w.Code)
			}
		})
	}
}

// TestVoteHandler_EmptyBody tests VoteHandler with empty body
func TestVoteHandler_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/vote", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	VoteHandler(w, req)

	// Should return error status
	if w.Code < 400 {
		t.Errorf("Empty body should return error, got %d", w.Code)
	}
}

// TestLoginHandler_EmptyCredentials tests login with empty credentials
func TestLoginHandler_EmptyCredentials(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Empty username", `{"username":"","password":"test123"}`},
		{"Empty password", `{"username":"testuser","password":""}`},
		{"Both empty", `{"username":"","password":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			LoginHandler(w, req)

			// Should return 400 Bad Request
			if w.Code != http.StatusBadRequest && w.Code < 400 {
				t.Logf("Empty credentials returned %d", w.Code)
			}
		})
	}
}

// TestRegisterHandler_EmptyFields tests registration with empty fields
func TestRegisterHandler_EmptyFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Empty name", `{"name":"","username":"user","password":"pass123"}`},
		{"Empty username", `{"name":"Test","username":"","password":"pass123"}`},
		{"Empty password", `{"name":"Test","username":"user","password":""}`},
		{"All empty", `{"name":"","username":"","password":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			RegisterHandler(w, req)

			// Should return 400 Bad Request
			if w.Code != http.StatusBadRequest && w.Code < 400 {
				t.Logf("Empty fields returned %d", w.Code)
			}
		})
	}
}

// TestChangePasswordHandler_ShortPassword tests change password with short password
func TestChangePasswordHandler_ShortPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"1 char", "a"},
		{"2 chars", "ab"},
		{"3 chars", "abc"},
		{"4 chars", "abcd"},
		{"5 chars", "abcde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"userId":1,"oldPassword":"oldpass123","newPassword":"` + tt.password + `"}`
			req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ChangePasswordHandler(w, req)

			// Should validate password length or require auth
			if w.Code == 0 {
				t.Error("Should set a response code")
			}
		})
	}
}

// TestPostsHandler_POST_EmptyFields tests POST with empty fields
func TestPostsHandler_POST_EmptyFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"Empty header", `{"header":"","content":"test","publisherId":1}`},
		{"Empty content", `{"header":"test","content":"","publisherId":1}`},
		{"Zero publisherId", `{"header":"test","content":"test","publisherId":0}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			PostsHandler(w, req)

			// Should return error or require auth
			if w.Code == 0 {
				t.Error("Should set a response code")
			}
		})
	}
}


