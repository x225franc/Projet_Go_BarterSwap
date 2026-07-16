package main

import (
	"errors"
	"net/http"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound            = errors.New("ressource introuvable")
	ErrInvalidInput        = errors.New("donnees invalides")
	ErrUnauthorized        = errors.New("authentification requise")
	ErrForbidden           = errors.New("action non autorisee")
	ErrConflict            = errors.New("conflit")
	ErrInsufficientCredits = errors.New("credits insuffisants")
	ErrInternal            = errors.New("erreur interne")
)

type domainError struct {
	sentinel error
	message  string
}

func (e *domainError) Error() string { return e.message }
func (e *domainError) Unwrap() error { return e.sentinel }

func newError(sentinel error, message string) error {
	return &domainError{sentinel: sentinel, message: message}
}

func statusForError(err error) (int, string) {
	message := "Erreur serveur"
	var de *domainError
	if errors.As(err, &de) {
		message = de.message
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, message
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInsufficientCredits):
		return http.StatusBadRequest, message
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, message
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, message
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, message
	default:
		return http.StatusInternalServerError, message
	}
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
