package main

import (
	"log"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/routes"
)

func main() {
	// Initialiser la base de données
	err := database.Initialize()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données: %v", err)
	}
	defer database.Close()

	// Configurer les routes
	router := routes.SetupRoutes()

	// Démarrer le serveur
	log.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
