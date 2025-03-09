package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Initialize initialise la connexion à la base de données et applique le schéma
func Initialize() error {
	// Vérifier si le fichier de base de données existe déjà
	_, err := os.Stat("forum.db")
	newDB := os.IsNotExist(err)

	// Ouvrir la connexion à la base de données
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return err
	}

	// Vérifier la connexion
	err = db.Ping()
	if err != nil {
		return err
	}

	DB = db

	// Si c'est une nouvelle base de données, appliquer le schéma
	if newDB {
		log.Println("Création d'une nouvelle base de données...")

		// Déterminer le chemin du fichier schema.sql
		// Chercher d'abord dans le répertoire courant
		schemaPath := "schema.sql"
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			// Essayer dans le répertoire parent
			schemaPath = filepath.Join("..", "schema.sql")
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				return err
			}
		}

		// Lire le fichier schema.sql
		schemaBytes, err := os.ReadFile(schemaPath)
		if err != nil {
			return err
		}

		// Exécuter les requêtes SQL
		_, err = DB.Exec(string(schemaBytes))
		if err != nil {
			return err
		}

		log.Println("Schéma de base de données initialisé avec succès")
	}

	return nil
}

// Close ferme la connexion à la base de données
func Close() {
	if DB != nil {
		DB.Close()
	}
}
