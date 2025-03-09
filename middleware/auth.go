package middleware

import (
	"context"
	"log"
	"net/http"
	"realtimeforum/database"
	"strings"
)

// Définir un type de clé pour le contexte
type contextKey string

// Clé pour stocker l'ID utilisateur dans le contexte
const UserIDKey contextKey = "userID"

// AuthMiddleware vérifie l'authentification de l'utilisateur
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Activer CORS pour les requêtes préliminaires OPTIONS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Répondre immédiatement aux requêtes OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Vérifier d'abord le cookie de session
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// Cookie non trouvé, vérifier l'en-tête Authorization
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				session, err := database.GetSessionByID(token)
				if err == nil {
					log.Printf("Authentification via token Bearer réussie pour l'utilisateur ID=%d", session.UserID)
					// Ajouter l'ID utilisateur au contexte
					ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else {
					log.Printf("Token Bearer invalide: %v", err)
				}
			} else {
				log.Printf("Aucun cookie ou token Bearer trouvé")
			}

			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}

		// Vérifier la session à partir du cookie
		session, err := database.GetSessionByID(cookie.Value)
		if err != nil {
			log.Printf("Session invalide: %v", err)
			http.Error(w, "Session invalide", http.StatusUnauthorized)
			return
		}

		log.Printf("Authentification via cookie réussie pour l'utilisateur ID=%d", session.UserID)

		// Ajouter l'ID utilisateur au contexte de la requête
		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)

		// Appeler le gestionnaire suivant avec le contexte mis à jour
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID récupère l'ID utilisateur à partir du contexte
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// OptionalAuthMiddleware permet l'accès que l'utilisateur soit authentifié ou non
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Activer CORS pour les requêtes préliminaires OPTIONS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Répondre immédiatement aux requêtes OPTIONS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Récupérer le cookie de session
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie != nil {
			// Vérifier la session
			session, err := database.GetSessionByID(cookie.Value)
			if err == nil && session != nil {
				// Ajouter l'ID utilisateur au contexte de la requête
				ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
				r = r.WithContext(ctx)
				log.Printf("Utilisateur authentifié (optionnel) ID=%d", session.UserID)
			}
		}

		// Appeler le gestionnaire suivant avec le contexte éventuellement mis à jour
		next.ServeHTTP(w, r)
	})
}

// WSAuthMiddleware vérifie l'authentification pour les connexions WebSocket
func WSAuthMiddleware(w http.ResponseWriter, r *http.Request) (int, bool) {
	// Activer CORS pour WebSocket
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Vérifier si le token est passé dans l'URL
	token := r.URL.Query().Get("token")
	if token != "" {
		session, err := database.GetSessionByID(token)
		if err == nil {
			log.Printf("Authentification WebSocket réussie via token URL pour l'utilisateur ID=%d", session.UserID)
			return session.UserID, true
		} else {
			log.Printf("Token URL invalide: %v", err)
		}
	}

	// Vérifier si un token est stocké dans localStorage et passé via la requête
	localStorageToken := r.URL.Query().Get("ls_token")
	if localStorageToken != "" {
		session, err := database.GetSessionByID(localStorageToken)
		if err == nil {
			log.Printf("Authentification WebSocket réussie via localStorage token pour l'utilisateur ID=%d", session.UserID)
			return session.UserID, true
		} else {
			log.Printf("localStorage token invalide: %v", err)
		}
	}

	// Vérifier le cookie de session
	cookie, err := r.Cookie("session_id")
	if err == nil {
		session, err := database.GetSessionByID(cookie.Value)
		if err == nil {
			log.Printf("Authentification WebSocket réussie via cookie pour l'utilisateur ID=%d", session.UserID)
			return session.UserID, true
		} else {
			log.Printf("Cookie de session invalide: %v", err)
		}
	}

	// Vérifier l'en-tête Authorization
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		session, err := database.GetSessionByID(token)
		if err == nil {
			log.Printf("Authentification WebSocket réussie via Authorization pour l'utilisateur ID=%d", session.UserID)
			return session.UserID, true
		} else {
			log.Printf("Token Bearer invalide: %v", err)
		}
	}

	log.Printf("Authentification WebSocket échouée: aucune méthode d'authentification valide")
	return 0, false
}
