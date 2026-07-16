package main

import (
	"context"
	"strconv"
)


func createExchange(ctx context.Context, requesterID string, serviceID int) (int, error) {
	ownerID, serviceCredits, err := dbFindActiveServiceForExchange(ctx, serviceID)
	if err != nil {
		return 0, newError(ErrNotFound, "Service introuvable ou inactif")
	}

	if strconv.Itoa(ownerID) == requesterID {
		return 0, newError(ErrInvalidInput, "Vous ne pouvez pas échanger un service avec vous-même")
	}

	userCredits, err := dbFindUserCreditBalance(ctx, requesterID)
	if err != nil || userCredits < serviceCredits {
		return 0, newError(ErrInvalidInput, "Crédits insuffisants pour lancer cet échange")
	}

	activeExchanges, err := dbCountActiveExchangesForService(ctx, serviceID)
	if err != nil {
		return 0, newError(ErrInternal, "Erreur serveur")
	}
	if activeExchanges > 0 {
		return 0, newError(ErrConflict, "Ce service est déjà en cours d'échange")
	}

	id, err := dbInsertExchange(ctx, serviceID, requesterID, ownerID)
	if err != nil {
		return 0, newError(ErrInternal, "Erreur création échange")
	}
	return id, nil
}

func listExchanges(ctx context.Context, userID, status string) ([]Exchange, error) {
	return dbFindExchangesForUser(ctx, userID, status)
}

func getExchange(ctx context.Context, id, userID string) (Exchange, error) {
	e, err := dbFindExchangeForUser(ctx, id, userID)
	if err != nil {
		return Exchange{}, newError(ErrNotFound, "Echange introuvable ou non autorisé")
	}
	return e, nil
}

func acceptExchange(ctx context.Context, id, ownerID string) error {
	requesterID, serviceCredits, status, err := dbFindExchangeForAccept(ctx, id, ownerID)
	if err != nil || status != "pending" {
		return newError(ErrInvalidInput, "Impossible d'accepter cet échange")
	}
	return dbAcceptExchangeTx(ctx, id, requesterID, serviceCredits)
}

func rejectExchange(ctx context.Context, id, ownerID string) error {
	rowsAffected, err := dbRejectExchangeForOwner(ctx, id, ownerID)
	if err != nil {
		return newError(ErrInternal, "Erreur serveur")
	}
	if rowsAffected == 0 {
		return newError(ErrInvalidInput, "Echange introuvable ou mauvais statut")
	}
	return nil
}

func completeExchange(ctx context.Context, id, requesterID string) error {
	ownerID, serviceCredits, err := dbFindExchangeForComplete(ctx, id, requesterID)
	if err != nil {
		return newError(ErrInvalidInput, "Echange introuvable ou non autorisé (seul le demandeur peut terminer l'échange)")
	}
	return dbCompleteExchangeTx(ctx, id, ownerID, serviceCredits)
}

func cancelExchange(ctx context.Context, id, userID string) (string, error) {
	requesterID, serviceCredits, status, err := dbFindExchangeForCancel(ctx, id, userID)
	if err != nil || (status != "pending" && status != "accepted") {
		return "", newError(ErrInvalidInput, "Impossible d'annuler cet échange (déjà terminé ou introuvable)")
	}

	if status == "accepted" {
		if err := dbCancelAcceptedExchangeTx(ctx, id, requesterID, serviceCredits); err != nil {
			return "", newError(ErrInternal, "Erreur lors du remboursement")
		}
		return "Echange annulé et crédits remboursés au demandeur", nil
	}

	if err := dbCancelPendingExchange(ctx, id); err != nil {
		return "", newError(ErrInternal, "Erreur lors de l'annulation")
	}
	return "Demande annulée avec succès", nil
}
