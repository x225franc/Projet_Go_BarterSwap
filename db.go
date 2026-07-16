package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

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
