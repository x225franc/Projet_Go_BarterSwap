package main

import (
	"context"
	"database/sql"
)


func dbScanReviews(ctx context.Context, query string, args ...any) ([]Review, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []Review{}
	for rows.Next() {
		var rev Review
		var comm sql.NullString
		var createdAt []byte

		if err := rows.Scan(&rev.ID, &rev.ExchangeID, &rev.AuthorID, &rev.TargetID, &rev.Note, &comm, &createdAt); err != nil {
			return nil, err
		}
		rev.Commentaire = comm.String
		rev.CreatedAt = string(createdAt)
		reviews = append(reviews, rev)
	}
	return reviews, rows.Err()
}

func dbFindExchangeForReview(ctx context.Context, exchangeID string) (requesterID, ownerID int, status string, err error) {
	err = db.QueryRowContext(ctx, "SELECT requester_id, owner_id, status FROM exchanges WHERE id = ?", exchangeID).
		Scan(&requesterID, &ownerID, &status)
	return
}

func dbInsertReview(ctx context.Context, exchangeID string, authorID, targetID int, rev Review) (int, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire) VALUES (?, ?, ?, ?, ?)",
		exchangeID, authorID, targetID, rev.Note, rev.Commentaire)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}
