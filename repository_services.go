package main

import (
	"context"
	"database/sql"
)

func dbInsertService(ctx context.Context, providerID string, s Service) (int, error) {
	query := `INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.ExecContext(ctx, query, providerID, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func dbFindServices(ctx context.Context, categorie, ville, search string) ([]Service, error) {
	query := `SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
	          FROM services WHERE 1=1`
	var args []any

	if categorie != "" {
		query += " AND categorie = ?"
		args = append(args, categorie)
	}
	if ville != "" {
		query += " AND ville = ?"
		args = append(args, ville)
	}
	if search != "" {
		query += " AND (titre LIKE ? OR description LIKE ?)"
		likeTerm := "%" + search + "%"
		args = append(args, likeTerm, likeTerm)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []Service{}
	for rows.Next() {
		var s Service
		var desc, v sql.NullString
		var createdAt []byte

		if err := rows.Scan(&s.ID, &s.ProviderID, &s.Titre, &desc, &s.Categorie, &s.DureeMinutes, &s.Credits, &v, &s.Actif, &createdAt); err != nil {
			return nil, err
		}
		s.Description = desc.String
		s.Ville = v.String
		s.CreatedAt = string(createdAt)
		services = append(services, s)
	}
	return services, rows.Err()
}

func dbFindServiceByID(ctx context.Context, id string) (Service, error) {
	var s Service
	var desc, ville sql.NullString
	var createdAt []byte

	query := `SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
	          FROM services WHERE id = ?`

	err := db.QueryRowContext(ctx, query, id).
		Scan(&s.ID, &s.ProviderID, &s.Titre, &desc, &s.Categorie, &s.DureeMinutes, &s.Credits, &ville, &s.Actif, &createdAt)
	if err != nil {
		return Service{}, err
	}

	s.Description = desc.String
	s.Ville = ville.String
	s.CreatedAt = string(createdAt)
	return s, nil
}

func dbUpdateServiceForProvider(ctx context.Context, id, providerID string, s Service) (int64, error) {
	query := `UPDATE services SET titre = ?, description = ?, categorie = ?, duree_minutes = ?, credits = ?, ville = ?, actif = ?
	          WHERE id = ? AND provider_id = ?`

	result, err := db.ExecContext(ctx, query, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif, id, providerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func dbDeleteServiceForProvider(ctx context.Context, id, providerID string) (int64, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM services WHERE id = ? AND provider_id = ?", id, providerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func dbFindServiceReviews(ctx context.Context, serviceID string) ([]Review, error) {
	query := `
		SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at
		FROM reviews r
		JOIN exchanges e ON r.exchange_id = e.id
		WHERE e.service_id = ? AND r.target_id = e.owner_id
	`
	return dbScanReviews(ctx, query, serviceID)
}
