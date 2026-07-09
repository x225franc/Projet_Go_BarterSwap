package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Routes Services
	mux.HandleFunc("POST /api/services", handleCreateService)
	mux.HandleFunc("GET /api/services", handleGetServices)
	mux.HandleFunc("GET /api/services/{id}", handleGetService)
	mux.HandleFunc("PUT /api/services/{id}", handleUpdateService)
	mux.HandleFunc("DELETE /api/services/{id}", handleDeleteService)

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