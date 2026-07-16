package main

import (
	"context"
	"database/sql"
)


func dbInsertUser(ctx context.Context, u User) (int, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (pseudo, bio, ville, credit_balance) VALUES (?, ?, ?, 10)",
		u.Pseudo, u.Bio, u.Ville)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func dbFindUserByID(ctx context.Context, id string) (User, error) {
	var u User
	var bio, ville sql.NullString
	var createdAt []byte

	err := db.QueryRowContext(ctx,
		"SELECT id, pseudo, bio, ville, credit_balance, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Pseudo, &bio, &ville, &u.CreditBalance, &createdAt)
	if err != nil {
		return User{}, err
	}

	u.Bio = bio.String
	u.Ville = ville.String
	u.CreatedAt = string(createdAt)
	return u, nil
}

func dbUpdateUserProfile(ctx context.Context, id string, u User) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET bio = ?, ville = ? WHERE id = ?", u.Bio, u.Ville, id)
	return err
}

func dbFindSkillsByUserID(ctx context.Context, userID string) ([]Skill, error) {
	skills := []Skill{}
	rows, err := db.QueryContext(ctx, "SELECT nom, niveau FROM skills WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

func dbReplaceUserSkills(ctx context.Context, userID string, skills []Skill) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM skills WHERE user_id = ?", userID); err != nil {
		return err
	}

	for _, skill := range skills {
		if _, err := tx.ExecContext(ctx, "INSERT INTO skills (user_id, nom, niveau) VALUES (?, ?, ?)", userID, skill.Nom, skill.Niveau); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func dbFindReviewsByTargetID(ctx context.Context, userID string) ([]Review, error) {
	return dbScanReviews(ctx,
		"SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at FROM reviews WHERE target_id = ?",
		userID)
}
