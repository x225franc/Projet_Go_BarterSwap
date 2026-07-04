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