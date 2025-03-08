// fichier: middleware/auth.go
package middleware

import (
	"context"
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
		// Récupérer le cookie de session
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// Pas de cookie trouvé, l'utilisateur n'est pas authentifié
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}

		// Vérifier la session
		session, err := database.GetSessionByID(cookie.Value)
		if err != nil {
			// Session invalide ou expirée
			http.Error(w, "Session invalide", http.StatusUnauthorized)
			return
		}

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
		// Récupérer le cookie de session
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie != nil {
			// Vérifier la session
			session, err := database.GetSessionByID(cookie.Value)
			if err == nil && session != nil {
				// Ajouter l'ID utilisateur au contexte de la requête
				ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
				r = r.WithContext(ctx)
			}
		}

		// Appeler le gestionnaire suivant avec le contexte éventuellement mis à jour
		next.ServeHTTP(w, r)
	})
}

// WSAuthMiddleware vérifie l'authentification pour les connexions WebSocket
func WSAuthMiddleware(w http.ResponseWriter, r *http.Request) (int, bool) {
	// Vérifier si le token est passé dans l'URL
	token := r.URL.Query().Get("token")
	if token != "" {
		session, err := database.GetSessionByID(token)
		if err == nil {
			return session.UserID, true
		}
	}

	// Vérifier le cookie de session
	cookie, err := r.Cookie("session_id")
	if err == nil {
		session, err := database.GetSessionByID(cookie.Value)
		if err == nil {
			return session.UserID, true
		}
	}

	// Vérifier l'en-tête Authorization
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		session, err := database.GetSessionByID(token)
		if err == nil {
			return session.UserID, true
		}
	}

	return 0, false
}
