package main

import "context"


func dbFindActiveServiceForExchange(ctx context.Context, serviceID int) (ownerID, credits int, err error) {
	err = db.QueryRowContext(ctx, "SELECT provider_id, credits FROM services WHERE id = ? AND actif = true", serviceID).
		Scan(&ownerID, &credits)
	return
}

func dbFindUserCreditBalance(ctx context.Context, userID string) (int, error) {
	var balance int
	err := db.QueryRowContext(ctx, "SELECT credit_balance FROM users WHERE id = ?", userID).Scan(&balance)
	return balance, err
}

func dbCountActiveExchangesForService(ctx context.Context, serviceID int) (int, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM exchanges WHERE service_id = ? AND status IN ('pending', 'accepted')", serviceID).
		Scan(&count)
	return count, err
}

func dbInsertExchange(ctx context.Context, serviceID int, requesterID string, ownerID int) (int, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO exchanges (service_id, requester_id, owner_id, status) VALUES (?, ?, ?, 'pending')",
		serviceID, requesterID, ownerID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func dbFindExchangesForUser(ctx context.Context, userID, status string) ([]Exchange, error) {
	query := "SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE (requester_id = ? OR owner_id = ?)"
	args := []any{userID, userID}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exchanges := []Exchange{}
	for rows.Next() {
		var e Exchange
		var ca, ua []byte
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &ca, &ua); err != nil {
			return nil, err
		}
		e.CreatedAt = string(ca)
		e.UpdatedAt = string(ua)
		exchanges = append(exchanges, e)
	}
	return exchanges, rows.Err()
}

func dbFindExchangeForUser(ctx context.Context, id, userID string) (Exchange, error) {
	var e Exchange
	var ca, ua []byte

	query := "SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE id = ? AND (requester_id = ? OR owner_id = ?)"
	err := db.QueryRowContext(ctx, query, id, userID, userID).
		Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &ca, &ua)
	if err != nil {
		return Exchange{}, err
	}

	e.CreatedAt = string(ca)
	e.UpdatedAt = string(ua)
	return e, nil
}

func dbFindExchangeForAccept(ctx context.Context, id, ownerID string) (requesterID, serviceCredits int, status string, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT e.requester_id, e.status, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND e.owner_id = ?",
		id, ownerID).Scan(&requesterID, &status, &serviceCredits)
	return
}

func dbAcceptExchangeTx(ctx context.Context, id string, requesterID, serviceCredits int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"UPDATE users SET credit_balance = credit_balance - ? WHERE id = ? AND credit_balance >= ?",
		serviceCredits, requesterID, serviceCredits)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		// Vérifié ici (et pas en amont) car c'est la seule façon d'éviter
		// une situation de concurrence entre la lecture et le débit du solde.
		return newError(ErrInsufficientCredits, "Le demandeur n'a plus assez de crédits")
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'spend')",
		requesterID, id, -serviceCredits); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE exchanges SET status = 'accepted' WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func dbRejectExchangeForOwner(ctx context.Context, id, ownerID string) (int64, error) {
	result, err := db.ExecContext(ctx,
		"UPDATE exchanges SET status = 'rejected' WHERE id = ? AND owner_id = ? AND status = 'pending'", id, ownerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func dbFindExchangeForComplete(ctx context.Context, id, requesterID string) (ownerID, serviceCredits int, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT e.owner_id, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND e.requester_id = ? AND e.status = 'accepted'",
		id, requesterID).Scan(&ownerID, &serviceCredits)
	return
}

func dbCompleteExchangeTx(ctx context.Context, id string, ownerID, serviceCredits int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?", serviceCredits, ownerID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'earn')",
		ownerID, id, serviceCredits); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE exchanges SET status = 'completed' WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func dbFindExchangeForCancel(ctx context.Context, id, userID string) (requesterID, serviceCredits int, status string, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT e.requester_id, e.status, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND (e.requester_id = ? OR e.owner_id = ?)",
		id, userID, userID).Scan(&requesterID, &status, &serviceCredits)
	return
}

func dbCancelAcceptedExchangeTx(ctx context.Context, id string, requesterID, serviceCredits int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?", serviceCredits, requesterID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'refund')",
		requesterID, id, serviceCredits); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE exchanges SET status = 'cancelled' WHERE id = ?", id); err != nil {
		return err
	}

	return tx.Commit()
}

func dbCancelPendingExchange(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "UPDATE exchanges SET status = 'cancelled' WHERE id = ?", id)
	return err
}
