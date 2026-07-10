package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// STRUCTURES DE DONNÉES

type User struct {
	ID            int     `json:"id"`
	Pseudo        string  `json:"pseudo"`
	Bio           string  `json:"bio,omitempty"`
	Ville         string  `json:"ville,omitempty"`
	Skills        []Skill `json:"skills"`
	CreditBalance int     `json:"credit_balance"`
	CreatedAt     string  `json:"created_at"`
}

type Skill struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}

type Service struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"` 
	Credits      int    `json:"credits"`       
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

type Exchange struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"` 
	OwnerID     int    `json:"owner_id"`     
	Status      string `json:"status"`       
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreditTransaction struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	ExchangeID int    `json:"exchange_id"`
	Montant    int    `json:"montant"` 
	Type       string `json:"type"`    
	CreatedAt  string `json:"created_at"`
}

type Review struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	AuthorID    int    `json:"author_id"`
	TargetID    int    `json:"target_id"`
	Note        int    `json:"note"` 
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// FONCTION PRINCIPALE

func main() {
	connectToDB()
	createTables()

	// config routes HTTP
	mux := http.NewServeMux()

	// Routes Utilisateurs
	mux.HandleFunc("POST /api/users", handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", handleUpdateUser)
	mux.HandleFunc("GET /api/users/{id}/skills", handleGetSkills)
	mux.HandleFunc("PUT /api/users/{id}/skills", handleUpdateSkills)
	mux.HandleFunc("GET /api/users/{id}/reviews", handleGetUserReviews)

	// Routes Services
	mux.HandleFunc("POST /api/services", handleCreateService)
	mux.HandleFunc("GET /api/services", handleGetServices)
	mux.HandleFunc("GET /api/services/{id}", handleGetService)
	mux.HandleFunc("PUT /api/services/{id}", handleUpdateService)
	mux.HandleFunc("DELETE /api/services/{id}", handleDeleteService)
	mux.HandleFunc("GET /api/services/{id}/reviews", handleGetServiceReviews)

	// Routes Echanges
	mux.HandleFunc("POST /api/exchanges", handleCreateExchange)
	mux.HandleFunc("GET /api/exchanges", handleGetExchanges)
	mux.HandleFunc("GET /api/exchanges/{id}", handleGetExchangeByID)
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", handleAcceptExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", handleRejectExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", handleCompleteExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", handleCancelExchange)
	mux.HandleFunc("POST /api/exchanges/{id}/review", handleCreateReview)

	fmt.Println("Serveur démarré sur http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}

// BDD

func connectToDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	dsn := fmt.Sprintf("barteruser:root@tcp(%s:3306)/barterswap_db", host)
	
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Impossible d'ouvrir la base de données:", err)
	}

	time.Sleep(3 * time.Second)

	err = db.Ping()
	if err != nil {
		log.Fatal("La base de données ne répond pas:", err)
	}
	fmt.Println("Connecté à MySQL avec succès !")
}

func createTables() {
	// Table users
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			pseudo VARCHAR(100) NOT NULL UNIQUE,
			bio TEXT,
			ville VARCHAR(100),
			credit_balance INT NOT NULL DEFAULT 10,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table users:", err)
	}

	// Table skills
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS skills (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			nom VARCHAR(100) NOT NULL,
			niveau VARCHAR(50) NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table skills:", err)
	}

	// Table services
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS services (
			id INT AUTO_INCREMENT PRIMARY KEY,
			provider_id INT NOT NULL,
			titre VARCHAR(255) NOT NULL,
			description TEXT,
			categorie VARCHAR(100) NOT NULL,
			duree_minutes INT NOT NULL,
			credits INT NOT NULL,
			ville VARCHAR(100),
			actif BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (provider_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table services:", err)
	}

	// Table exchanges
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS exchanges (
			id INT AUTO_INCREMENT PRIMARY KEY,
			service_id INT NOT NULL,
			requester_id INT NOT NULL,
			owner_id INT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (service_id) REFERENCES services(id),
			FOREIGN KEY (requester_id) REFERENCES users(id),
			FOREIGN KEY (owner_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table exchanges:", err)
	}

	// Table credit_transactions
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS credit_transactions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			exchange_id INT NOT NULL,
			montant INT NOT NULL,
			type VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (exchange_id) REFERENCES exchanges(id)
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table credit_transactions:", err)
	}

	// Table reviews
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reviews (
			id INT AUTO_INCREMENT PRIMARY KEY,
			exchange_id INT NOT NULL,
			author_id INT NOT NULL,
			target_id INT NOT NULL,
			note INT NOT NULL CHECK(note >= 1 AND note <= 5),
			commentaire TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (exchange_id) REFERENCES exchanges(id),
			FOREIGN KEY (author_id) REFERENCES users(id),
			FOREIGN KEY (target_id) REFERENCES users(id),
			UNIQUE KEY unique_review (exchange_id, author_id)
		)
	`)
	if err != nil {
		log.Fatal("Erreur création table reviews:", err)
	}
}

// HANDLERS HTTP UTILISATEURS

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil || u.Pseudo == "" {
		http.Error(w, `{"error": "Données invalides ou pseudo manquant"}`, http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO users (pseudo, bio, ville, credit_balance) VALUES (?, ?, ?, 10)", u.Pseudo, u.Bio, u.Ville)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur ou pseudo déjà pris"}`, http.StatusInternalServerError)
		return
	}

	newID, _ := result.LastInsertId()
	u.ID = int(newID)
	u.CreditBalance = 10

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var u User
	var bio, ville sql.NullString
	var createdAt []byte

	err := db.QueryRow("SELECT id, pseudo, bio, ville, credit_balance, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Pseudo, &bio, &ville, &u.CreditBalance, &createdAt)
	
	if err != nil {
		http.Error(w, `{"error": "Utilisateur introuvable"}`, http.StatusNotFound)
		return
	}

	u.Bio = bio.String
	u.Ville = ville.String
	u.CreatedAt = string(createdAt)

	u.Skills = getSkillsForUser(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	if r.Header.Get("X-User-ID") != id {
		http.Error(w, `{"error": "Non autorisé"}`, http.StatusUnauthorized)
		return
	}

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, `{"error": "Données invalides"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET bio = ?, ville = ? WHERE id = ?", u.Bio, u.Ville, id)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la mise à jour"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Profil mis à jour"}`))
}

func handleGetSkills(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skills := getSkillsForUser(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(skills)
}

func handleUpdateSkills(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	if r.Header.Get("X-User-ID") != id {
		http.Error(w, `{"error": "Non autorisé"}`, http.StatusUnauthorized)
		return
	}

	var skills []Skill
	err := json.NewDecoder(r.Body).Decode(&skills)
	if err != nil {
		http.Error(w, `{"error": "Données invalides"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM skills WHERE user_id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}

	for _, skill := range skills {
		_, err = db.Exec("INSERT INTO skills (user_id, nom, niveau) VALUES (?, ?, ?)", id, skill.Nom, skill.Niveau)
		if err != nil {
			http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Compétences mises à jour"}`))
}

func handleGetUserReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	query := "SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at FROM reviews WHERE target_id = ?"
	rows, err := db.Query(query, id)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []Review = []Review{}
	for rows.Next() {
		var rev Review
		var comm sql.NullString
		var ca []byte
		rows.Scan(&rev.ID, &rev.ExchangeID, &rev.AuthorID, &rev.TargetID, &rev.Note, &comm, &ca)
		rev.Commentaire = comm.String
		rev.CreatedAt = string(ca)
		reviews = append(reviews, rev)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// HANDLERS HTTP SERVICES

func handleCreateService(w http.ResponseWriter, r *http.Request) {
	providerID := r.Header.Get("X-User-ID")
	if providerID == "" {
		http.Error(w, `{"error": "Authentification requise (X-User-ID manquant)"}`, http.StatusUnauthorized)
		return
	}

	var s Service
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil || s.Titre == "" || s.Categorie == "" {
		http.Error(w, `{"error": "Données invalides (titre et catégorie obligatoires)"}`, http.StatusBadRequest)
		return
	}

	if !isValidCategory(s.Categorie) {
		http.Error(w, `{"error": "Catégorie non reconnue"}`, http.StatusBadRequest)
		return
	}

	query := `INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	s.Actif = true

	result, err := db.Exec(query, providerID, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la création du service"}`, http.StatusInternalServerError)
		return
	}

	newID, _ := result.LastInsertId()
	s.ID = int(newID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func handleGetServices(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at 
	          FROM services WHERE 1=1`
	var args []any

	cat := r.URL.Query().Get("categorie")
	if cat != "" {
		query += " AND categorie = ?"
		args = append(args, cat)
	}

	ville := r.URL.Query().Get("ville")
	if ville != "" {
		query += " AND ville = ?"
		args = append(args, ville)
	}

	search := r.URL.Query().Get("search")
	if search != "" {
		query += " AND (titre LIKE ? OR description LIKE ?)"
		likeTerm := "%" + search + "%"
		args = append(args, likeTerm, likeTerm)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la récupération des services"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		var desc, ville sql.NullString
		var createdAt []byte

		err := rows.Scan(&s.ID, &s.ProviderID, &s.Titre, &desc, &s.Categorie, &s.DureeMinutes, &s.Credits, &ville, &s.Actif, &createdAt)
		if err == nil {
			s.Description = desc.String
			s.Ville = ville.String
			s.CreatedAt = string(createdAt)
			services = append(services, s)
		}
	}

	if services == nil {
		services = []Service{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func handleGetService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var s Service
	var desc, ville sql.NullString
	var createdAt []byte

	query := `SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at 
	          FROM services WHERE id = ?`

	err := db.QueryRow(query, id).Scan(&s.ID, &s.ProviderID, &s.Titre, &desc, &s.Categorie, &s.DureeMinutes, &s.Credits, &ville, &s.Actif, &createdAt)
	
	if err != nil {
		http.Error(w, `{"error": "Service introuvable"}`, http.StatusNotFound)
		return
	}

	s.Description = desc.String
	s.Ville = ville.String
	s.CreatedAt = string(createdAt)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func handleUpdateService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	providerID := r.Header.Get("X-User-ID")

	var s Service
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, `{"error": "Données invalides"}`, http.StatusBadRequest)
		return
	}

	query := `UPDATE services SET titre = ?, description = ?, categorie = ?, duree_minutes = ?, credits = ?, ville = ?, actif = ? 
	          WHERE id = ? AND provider_id = ?`

	result, err := db.Exec(query, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif, id, providerID)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "Non autorisé ou service introuvable"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Service mis à jour avec succès"}`))
}

func handleDeleteService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	providerID := r.Header.Get("X-User-ID")

	query := "DELETE FROM services WHERE id = ? AND provider_id = ?"
	result, err := db.Exec(query, id, providerID)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "Non autorisé ou service introuvable"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Service supprimé avec succès"}`))
}

func handleGetServiceReviews(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	query := `
		SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at 
		FROM reviews r
		JOIN exchanges e ON r.exchange_id = e.id
		WHERE e.service_id = ? AND r.target_id = e.owner_id
	`
	rows, err := db.Query(query, id)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviews []Review = []Review{}
	for rows.Next() {
		var rev Review
		var comm sql.NullString
		var ca []byte
		rows.Scan(&rev.ID, &rev.ExchangeID, &rev.AuthorID, &rev.TargetID, &rev.Note, &comm, &ca)
		rev.Commentaire = comm.String
		rev.CreatedAt = string(ca)
		reviews = append(reviews, rev)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// HANDLERS HTTP ECHANGES

func handleCreateExchange(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Header.Get("X-User-ID")
	if requesterID == "" {
		http.Error(w, `{"error": "Authentification requise"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		ServiceID int `json:"service_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Données invalides"}`, http.StatusBadRequest)
		return
	}

	var ownerID int
	var serviceCredits int
	err := db.QueryRow("SELECT provider_id, credits FROM services WHERE id = ? AND actif = true", req.ServiceID).Scan(&ownerID, &serviceCredits)
	if err != nil {
		http.Error(w, `{"error": "Service introuvable ou inactif"}`, http.StatusNotFound)
		return
	}

	if fmt.Sprintf("%d", ownerID) == requesterID {
		http.Error(w, `{"error": "Vous ne pouvez pas échanger un service avec vous-même"}`, http.StatusBadRequest)
		return
	}

	var userCredits int
	err = db.QueryRow("SELECT credit_balance FROM users WHERE id = ?", requesterID).Scan(&userCredits)
	if err != nil || userCredits < serviceCredits {
		http.Error(w, `{"error": "Crédits insuffisants pour lancer cet échange"}`, http.StatusBadRequest)
		return
	}

	var activeExchanges int
	db.QueryRow("SELECT COUNT(*) FROM exchanges WHERE service_id = ? AND status IN ('pending', 'accepted')", req.ServiceID).Scan(&activeExchanges)
	if activeExchanges > 0 {
		http.Error(w, `{"error": "Ce service est déjà en cours d'échange"}`, http.StatusConflict)
		return
	}

	query := "INSERT INTO exchanges (service_id, requester_id, owner_id, status) VALUES (?, ?, ?, 'pending')"
	res, err := db.Exec(query, req.ServiceID, requesterID, ownerID)
	if err != nil {
		http.Error(w, `{"error": "Erreur création échange"}`, http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"message": "Demande créée", "exchange_id": %d}`, id)))
}

func handleGetExchanges(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	status := r.URL.Query().Get("status")

	query := "SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE (requester_id = ? OR owner_id = ?)"
	args := []any{userID, userID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var exchanges []Exchange = []Exchange{}
	for rows.Next() {
		var e Exchange
		var ca, ua []byte
		rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &ca, &ua)
		e.CreatedAt = string(ca)
		e.UpdatedAt = string(ua)
		exchanges = append(exchanges, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exchanges)
}

func handleGetExchangeByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")

	var e Exchange
	var ca, ua []byte
	query := "SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE id = ? AND (requester_id = ? OR owner_id = ?)"
	
	err := db.QueryRow(query, id, userID, userID).Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &ca, &ua)
	if err != nil {
		http.Error(w, `{"error": "Echange introuvable ou non autorisé"}`, http.StatusNotFound)
		return
	}

	e.CreatedAt = string(ca)
	e.UpdatedAt = string(ua)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func handleAcceptExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ownerID := r.Header.Get("X-User-ID")

	var e Exchange
	var serviceCredits int
	err := db.QueryRow("SELECT e.requester_id, e.status, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND e.owner_id = ?", id, ownerID).Scan(&e.RequesterID, &e.Status, &serviceCredits)
	
	if err != nil || e.Status != "pending" {
		http.Error(w, `{"error": "Impossible d'accepter cet échange"}`, http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE users SET credit_balance = credit_balance - ? WHERE id = ? AND credit_balance >= ?", serviceCredits, e.RequesterID, serviceCredits)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la mise à jour des crédits"}`, http.StatusInternalServerError)
		return
	}
	
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "Le demandeur n'a plus assez de crédits"}`, http.StatusBadRequest)
		return
	}

	_, err = tx.Exec("INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'spend')", e.RequesterID, id, -serviceCredits)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de l'enregistrement de la transaction"}`, http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE exchanges SET status = 'accepted' WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de la mise à jour du statut"}`, http.StatusInternalServerError)
		return
	}

	tx.Commit()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Echange accepté, crédits bloqués"}`))
}

func handleRejectExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ownerID := r.Header.Get("X-User-ID")

	res, err := db.Exec("UPDATE exchanges SET status = 'rejected' WHERE id = ? AND owner_id = ? AND status = 'pending'", id, ownerID)
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"error": "Echange introuvable ou mauvais statut"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Echange refusé"}`))
}

func handleCompleteExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	requesterID := r.Header.Get("X-User-ID")

	var ownerID int
	var serviceCredits int
	err := db.QueryRow("SELECT e.owner_id, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND e.requester_id = ? AND e.status = 'accepted'", id, requesterID).Scan(&ownerID, &serviceCredits)
	
	if err != nil {
		http.Error(w, `{"error": "Echange introuvable ou non autorisé (seul le demandeur peut terminer l'échange)"}`, http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?", serviceCredits, ownerID)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors du transfert de crédits"}`, http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'earn')", ownerID, id, serviceCredits)
	if err != nil {
		http.Error(w, `{"error": "Erreur trace de transaction"}`, http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE exchanges SET status = 'completed' WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Erreur mise à jour du statut"}`, http.StatusInternalServerError)
		return
	}

	tx.Commit()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Echange terminé, crédits transférés"}`))
}

func handleCancelExchange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")

	var e Exchange
	var serviceCredits int
	err := db.QueryRow("SELECT e.requester_id, e.status, s.credits FROM exchanges e JOIN services s ON e.service_id = s.id WHERE e.id = ? AND (e.requester_id = ? OR e.owner_id = ?)", id, userID, userID).Scan(&e.RequesterID, &e.Status, &serviceCredits)
	
	if err != nil || (e.Status != "pending" && e.Status != "accepted") {
		http.Error(w, `{"error": "Impossible d'annuler cet échange (déjà terminé ou introuvable)"}`, http.StatusBadRequest)
		return
	}

	if e.Status == "accepted" {
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, `{"error": "Erreur serveur"}`, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("UPDATE users SET credit_balance = credit_balance + ? WHERE id = ?", serviceCredits, e.RequesterID)
		if err != nil {
			http.Error(w, `{"error": "Erreur lors du remboursement"}`, http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES (?, ?, ?, 'refund')", e.RequesterID, id, serviceCredits)
		if err != nil {
			http.Error(w, `{"error": "Erreur trace de transaction"}`, http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("UPDATE exchanges SET status = 'cancelled' WHERE id = ?", id)
		if err != nil {
			http.Error(w, `{"error": "Erreur mise à jour statut"}`, http.StatusInternalServerError)
			return
		}

		tx.Commit()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Echange annulé et crédits remboursés au demandeur"}`))
		return
	}

	_, err = db.Exec("UPDATE exchanges SET status = 'cancelled' WHERE id = ?", id)
	if err != nil {
		http.Error(w, `{"error": "Erreur lors de l'annulation"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Demande annulée avec succès"}`))
}

func handleCreateReview(w http.ResponseWriter, r *http.Request) {
	exchangeID := r.PathValue("id")
	authorIDStr := r.Header.Get("X-User-ID")
	authorID, _ := strconv.Atoi(authorIDStr)

	if authorIDStr == "" {
		http.Error(w, `{"error": "Authentification requise"}`, http.StatusUnauthorized)
		return
	}

	var rev Review
	err := json.NewDecoder(r.Body).Decode(&rev)
	if err != nil || rev.Note < 1 || rev.Note > 5 {
		http.Error(w, `{"error": "Données invalides ou note hors limite (doit être entre 1 et 5)"}`, http.StatusBadRequest)
		return
	}

	var reqID, ownerID int
	var status string
	err = db.QueryRow("SELECT requester_id, owner_id, status FROM exchanges WHERE id = ?", exchangeID).Scan(&reqID, &ownerID, &status)
	if err != nil {
		http.Error(w, `{"error": "Echange introuvable"}`, http.StatusNotFound)
		return
	}

	if status != "completed" {
		http.Error(w, `{"error": "Vous ne pouvez évaluer qu'un échange terminé"}`, http.StatusBadRequest)
		return
	}

	if authorID != reqID && authorID != ownerID {
		http.Error(w, `{"error": "Vous n'êtes pas impliqué dans cet échange"}`, http.StatusForbidden)
		return
	}

	targetID := ownerID
	if authorID == ownerID {
		targetID = reqID
	}

	query := "INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire) VALUES (?, ?, ?, ?, ?)"
	res, err := db.Exec(query, exchangeID, authorID, targetID, rev.Note, rev.Commentaire)
	if err != nil {
		http.Error(w, `{"error": "Vous avez déjà évalué cet échange ou erreur serveur"}`, http.StatusConflict)
		return
	}

	newID, _ := res.LastInsertId()
	rev.ID = int(newID)
	rev.ExchangeID, _ = strconv.Atoi(exchangeID)
	rev.AuthorID = authorID
	rev.TargetID = targetID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rev)
}

// FONCTIONS UTILITAIRES

func getSkillsForUser(userID string) []Skill {
	skills := []Skill{}
	rows, err := db.Query("SELECT nom, niveau FROM skills WHERE user_id = ?", userID)
	if err != nil {
		return skills
	}
	defer rows.Close()

	for rows.Next() {
		var s Skill
		rows.Scan(&s.Nom, &s.Niveau)
		skills = append(skills, s)
	}
	return skills
}

func isValidCategory(c string) bool {
	categories := []string{
		"Informatique", "Jardinage", "Bricolage", "Cuisine", "Musique", 
		"Langues", "Sport", "Tutorat", "Déménagement", "Photographie", 
		"Animalier", "Couture", "Autre",
	}
	for _, cat := range categories {
		if c == cat {
			return true
		}
	}
	return false
}