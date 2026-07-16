package main

import (
	"context"
	"strconv"
)


func createReview(ctx context.Context, exchangeIDStr, authorIDStr string, rev Review) (Review, error) {
	authorID, _ := strconv.Atoi(authorIDStr)

	if rev.Note < 1 || rev.Note > 5 {
		return Review{}, newError(ErrInvalidInput, "Données invalides ou note hors limite (doit être entre 1 et 5)")
	}

	reqID, ownerID, status, err := dbFindExchangeForReview(ctx, exchangeIDStr)
	if err != nil {
		return Review{}, newError(ErrNotFound, "Echange introuvable")
	}
	if status != "completed" {
		return Review{}, newError(ErrInvalidInput, "Vous ne pouvez évaluer qu'un échange terminé")
	}
	if authorID != reqID && authorID != ownerID {
		return Review{}, newError(ErrForbidden, "Vous n'êtes pas impliqué dans cet échange")
	}

	targetID := ownerID
	if authorID == ownerID {
		targetID = reqID
	}

	id, err := dbInsertReview(ctx, exchangeIDStr, authorID, targetID, rev)
	if err != nil {
		return Review{}, newError(ErrConflict, "Vous avez déjà évalué cet échange ou erreur serveur")
	}

	rev.ID = id
	rev.ExchangeID, _ = strconv.Atoi(exchangeIDStr)
	rev.AuthorID = authorID
	rev.TargetID = targetID
	return rev, nil
}
