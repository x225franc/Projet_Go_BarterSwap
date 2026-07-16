package main

import (
	"context"
	"database/sql"
	"errors"
)


var validCategories = []string{
	"Informatique", "Jardinage", "Bricolage", "Cuisine", "Musique",
	"Langues", "Sport", "Tutorat", "Déménagement", "Photographie",
	"Animalier", "Couture", "Autre",
}

func isValidCategory(c string) bool {
	for _, cat := range validCategories {
		if c == cat {
			return true
		}
	}
	return false
}

func createService(ctx context.Context, providerID string, s Service) (Service, error) {
	if s.Titre == "" || s.Categorie == "" {
		return Service{}, newError(ErrInvalidInput, "Données invalides (titre et catégorie obligatoires)")
	}
	if !isValidCategory(s.Categorie) {
		return Service{}, newError(ErrInvalidInput, "Catégorie non reconnue")
	}

	s.Actif = true
	id, err := dbInsertService(ctx, providerID, s)
	if err != nil {
		return Service{}, newError(ErrInternal, "Erreur lors de la création du service")
	}
	s.ID = id

	return s, nil
}

func listServices(ctx context.Context, categorie, ville, search string) ([]Service, error) {
	return dbFindServices(ctx, categorie, ville, search)
}

func getService(ctx context.Context, id string) (Service, error) {
	s, err := dbFindServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Service{}, newError(ErrNotFound, "Service introuvable")
		}
		return Service{}, newError(ErrInternal, "Erreur serveur")
	}
	return s, nil
}

func updateService(ctx context.Context, id, providerID string, s Service) error {
	rowsAffected, err := dbUpdateServiceForProvider(ctx, id, providerID, s)
	if err != nil {
		return newError(ErrInternal, "Erreur serveur")
	}
	if rowsAffected == 0 {
		return newError(ErrForbidden, "Non autorisé ou service introuvable")
	}
	return nil
}

func deleteService(ctx context.Context, id, providerID string) error {
	rowsAffected, err := dbDeleteServiceForProvider(ctx, id, providerID)
	if err != nil {
		return newError(ErrInternal, "Erreur serveur")
	}
	if rowsAffected == 0 {
		return newError(ErrForbidden, "Non autorisé ou service introuvable")
	}
	return nil
}

func getServiceReviews(ctx context.Context, id string) ([]Review, error) {
	return dbFindServiceReviews(ctx, id)
}
