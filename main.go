package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	connectToDB()
	createTables()

	fmt.Println("Serveur démarré sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", newRouter()); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
