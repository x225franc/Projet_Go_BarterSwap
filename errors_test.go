package main

import (
	"errors"
	"net/http"
	"testing"
)

func TestNewErrorWrapsSentinel(t *testing.T) {
	err := newError(ErrNotFound, "utilisateur introuvable")

	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is devrait reconnaître la sentinelle enveloppée")
	}
	if errors.Is(err, ErrConflict) {
		t.Error("errors.Is ne devrait pas matcher une sentinelle différente")
	}
	if err.Error() != "utilisateur introuvable" {
		t.Errorf("Error() = %q, want %q", err.Error(), "utilisateur introuvable")
	}
}

func TestStatusForError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", newError(ErrNotFound, "x"), http.StatusNotFound},
		{"invalid input", newError(ErrInvalidInput, "x"), http.StatusBadRequest},
		{"insufficient credits", newError(ErrInsufficientCredits, "x"), http.StatusBadRequest},
		{"unauthorized", newError(ErrUnauthorized, "x"), http.StatusUnauthorized},
		{"forbidden", newError(ErrForbidden, "x"), http.StatusForbidden},
		{"conflict", newError(ErrConflict, "x"), http.StatusConflict},
		{"internal", newError(ErrInternal, "x"), http.StatusInternalServerError},
		{"erreur brute non enveloppée", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := statusForError(tc.err)
			if status != tc.wantStatus {
				t.Errorf("statusForError(%v) status = %d, want %d", tc.err, status, tc.wantStatus)
			}
		})
	}
}

func TestStatusForErrorMessage(t *testing.T) {
	err := newError(ErrInvalidInput, "message précis pour l'utilisateur")
	_, message := statusForError(err)
	if message != "message précis pour l'utilisateur" {
		t.Errorf("message = %q, want le message d'origine", message)
	}
}

func TestIsDuplicateEntryOnPlainError(t *testing.T) {
	if isDuplicateEntry(errors.New("erreur quelconque")) {
		t.Error("une erreur générique ne doit jamais être considérée comme un doublon")
	}
	if isDuplicateEntry(nil) {
		t.Error("nil ne doit pas être considéré comme un doublon")
	}
}
