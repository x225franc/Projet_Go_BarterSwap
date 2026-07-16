package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"message": "ok"})

	if rec.Code != http.StatusCreated {
		t.Errorf("code = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("réponse non décodable: %v", err)
	}
	if body["message"] != "ok" {
		t.Errorf("message = %q, want ok", body["message"])
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, newError(ErrForbidden, "accès refusé"))

	if rec.Code != http.StatusForbidden {
		t.Errorf("code = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("réponse non décodable: %v", err)
	}
	if body["error"] != "accès refusé" {
		t.Errorf("error = %q, want %q", body["error"], "accès refusé")
	}
}

func TestDecodeJSONValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"pseudo":"alice"}`))
	var u User
	if err := decodeJSON(req, &u); err != nil {
		t.Fatalf("decodeJSON a échoué sur un JSON valide: %v", err)
	}
	if u.Pseudo != "alice" {
		t.Errorf("Pseudo = %q, want alice", u.Pseudo)
	}
}

func TestDecodeJSONInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalide`))
	var u User
	err := decodeJSON(req, &u)
	if err == nil {
		t.Fatal("decodeJSON aurait dû échouer sur un JSON invalide")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("l'erreur devrait envelopper ErrInvalidInput, got %v", err)
	}
}

func TestIsOwner(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/42", nil)
	req.Header.Set("X-User-ID", "42")

	if !isOwner(req, "42") {
		t.Error("isOwner devrait renvoyer true quand le header correspond à l'id")
	}
	if isOwner(req, "43") {
		t.Error("isOwner devrait renvoyer false quand le header ne correspond pas à l'id")
	}
}

func TestIsOwnerMissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/42", nil)
	if isOwner(req, "42") {
		t.Error("isOwner devrait renvoyer false quand le header X-User-ID est absent")
	}
}
