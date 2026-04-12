package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"app-backend/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 8 {
		writeJSONError(w, http.StatusBadRequest, "email is required and password must be at least 8 characters")
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrEmailInUse) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	token, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			writeJSONError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
