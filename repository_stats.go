package main

import "context"


func dbCountActiveServicesForProvider(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM services WHERE provider_id = ? AND actif = true", userID).
		Scan(&count)
	return count, err
}

func dbCountCompletedExchangesForUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM exchanges WHERE (requester_id = ? OR owner_id = ?) AND status = 'completed'",
		userID, userID).Scan(&count)
	return count, err
}

func dbReviewStatsForUser(ctx context.Context, userID string) (noteMoyenne float64, nbAvis int, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT COALESCE(AVG(note), 0), COUNT(*) FROM reviews WHERE target_id = ?", userID).
		Scan(&noteMoyenne, &nbAvis)
	return
}

func dbCreditTotalsForUser(ctx context.Context, userID string) (totalGagne, totalDepense int, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'earn' THEN montant ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'spend' THEN -montant ELSE 0 END), 0)
		FROM credit_transactions WHERE user_id = ?
	`, userID).Scan(&totalGagne, &totalDepense)
	return
}
