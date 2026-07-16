package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

func createUser(ctx context.Context, u User) (User, error) {
	if u.Pseudo == "" {
		return User{}, newError(ErrInvalidInput, "Données invalides ou pseudo manquant")
	}

	id, err := dbInsertUser(ctx, u)
	if err != nil {
		if isDuplicateEntry(err) {
			return User{}, newError(ErrConflict, "Ce pseudo est déjà pris")
		}
		return User{}, newError(ErrInternal, "Erreur serveur")
	}

	created, err := dbFindUserByID(ctx, strconv.Itoa(id))
	if err != nil {
		return User{}, newError(ErrInternal, "Erreur serveur")
	}
	created.Skills = []Skill{}
	return created, nil
}

func getUserProfile(ctx context.Context, id string) (User, error) {
	u, err := dbFindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, newError(ErrNotFound, "Utilisateur introuvable")
		}
		return User{}, newError(ErrInternal, "Erreur serveur")
	}

	skills, err := dbFindSkillsByUserID(ctx, id)
	if err != nil {
		return User{}, newError(ErrInternal, "Erreur serveur")
	}
	u.Skills = skills

	return u, nil
}

func updateUserProfile(ctx context.Context, id string, u User) error {
	if err := dbUpdateUserProfile(ctx, id, u); err != nil {
		return newError(ErrInternal, "Erreur lors de la mise à jour")
	}
	return nil
}

func getUserSkills(ctx context.Context, id string) ([]Skill, error) {
	return dbFindSkillsByUserID(ctx, id)
}

func setUserSkills(ctx context.Context, id string, skills []Skill) error {
	if err := dbReplaceUserSkills(ctx, id, skills); err != nil {
		return newError(ErrInternal, "Erreur serveur")
	}
	return nil
}

func getUserReviews(ctx context.Context, id string) ([]Review, error) {
	return dbFindReviewsByTargetID(ctx, id)
}

func getUserStats(ctx context.Context, id string) (UserStats, error) {
	u, err := dbFindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserStats{}, newError(ErrNotFound, "Utilisateur introuvable")
		}
		return UserStats{}, newError(ErrInternal, "Erreur serveur")
	}

	servicesActifs, err := dbCountActiveServicesForProvider(ctx, id)
	if err != nil {
		return UserStats{}, newError(ErrInternal, "Erreur serveur")
	}

	echangesCompletes, err := dbCountCompletedExchangesForUser(ctx, id)
	if err != nil {
		return UserStats{}, newError(ErrInternal, "Erreur serveur")
	}

	noteMoyenne, nbAvis, err := dbReviewStatsForUser(ctx, id)
	if err != nil {
		return UserStats{}, newError(ErrInternal, "Erreur serveur")
	}

	totalGagne, totalDepense, err := dbCreditTotalsForUser(ctx, id)
	if err != nil {
		return UserStats{}, newError(ErrInternal, "Erreur serveur")
	}

	return UserStats{
		UserID:            u.ID,
		ServicesActifs:    servicesActifs,
		EchangesCompletes: echangesCompletes,
		CreditBalance:     u.CreditBalance,
		NoteMoyenne:       noteMoyenne,
		NbAvis:            nbAvis,
		TotalGagne:        totalGagne,
		TotalDepense:      totalDepense,
	}, nil
}
