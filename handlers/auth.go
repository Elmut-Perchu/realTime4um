// fichier: handlers/auth.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/middleware"
	"strings"
	"time"
)

// RegisterHandler gère l'inscription d'un nouvel utilisateur
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Activer CORS pour cette réponse
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Décoder le corps de la requête
	var userDTO database.UserDTO
	err := json.NewDecoder(r.Body).Decode(&userDTO)
	if err != nil {
		log.Printf("Erreur de décodage du corps de la requête: %v", err)
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if userDTO.Username == "" || userDTO.Email == "" || userDTO.Password == "" ||
		userDTO.FirstName == "" || userDTO.LastName == "" || userDTO.Age < 13 ||
		(userDTO.Gender != "M" && userDTO.Gender != "F" && userDTO.Gender != "Autre") {
		http.Error(w, "Données incomplètes ou invalides", http.StatusBadRequest)
		return
	}

	// Créer l'utilisateur
	log.Printf("Tentative de création d'utilisateur: %s, %s", userDTO.Username, userDTO.Email)
	userID, err := database.CreateUser(userDTO)
	if err != nil {
		log.Printf("Erreur lors de la création de l'utilisateur: %v", err)
		// Vérifier si l'erreur est liée à une contrainte d'unicité
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "users.email") {
				http.Error(w, "Erreur lors de la création de l'utilisateur: Email déjà utilisé", http.StatusConflict)
			} else if strings.Contains(err.Error(), "users.username") {
				http.Error(w, "Erreur lors de la création de l'utilisateur: Nom d'utilisateur déjà utilisé", http.StatusConflict)
			} else {
				http.Error(w, "Erreur lors de la création de l'utilisateur: "+err.Error(), http.StatusConflict)
			}
		} else {
			http.Error(w, "Erreur lors de la création de l'utilisateur: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	log.Printf("Utilisateur créé avec l'ID: %d", userID)

	// Créer une session pour l'utilisateur
	log.Printf("Tentative de création de session pour l'utilisateur: %d", userID)
	session, err := database.CreateSession(userID)
	if err != nil {
		log.Printf("Erreur lors de la création de la session: %v", err)
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}
	log.Printf("Session créée avec succès: %s", session.ID)

	// Définir le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: false, // Permettre l'accès via JavaScript
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Secure:   false, // Ne pas exiger HTTPS en développement
	})

	// Ajouter l'ID de session dans l'en-tête pour le client JavaScript
	w.Header().Set("X-Session-ID", session.ID)

	// Récupérer l'utilisateur créé
	log.Printf("Récupération de l'utilisateur avec l'ID: %d", userID)
	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("Erreur lors de la récupération de l'utilisateur: %v", err)
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}
	log.Printf("Utilisateur récupéré avec succès: %s", user.Username)

	// Nettoyer le mot de passe avant de retourner l'utilisateur
	user.Password = ""

	// Créer une réponse personnalisée avec l'utilisateur et l'ID de session
	response := struct {
		*database.User
		SessionID string `json:"sessionId"`
	}{
		User:      user,
		SessionID: session.ID,
	}

	// Définir le type de contenu avant d'écrire quoi que ce soit
	w.Header().Set("Content-Type", "application/json")
	// Écrire le code de statut explicitement
	w.WriteHeader(http.StatusCreated)

	// Sérialiser manuellement pour éviter les problèmes
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de la réponse: %v", err)
		http.Error(w, "Erreur lors de la sérialisation de la réponse", http.StatusInternalServerError)
		return
	}

	// Écrire directement dans le ResponseWriter
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Printf("Erreur lors de l'écriture de la réponse: %v", err)
	}
}

// LoginHandler gère la connexion d'un utilisateur
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Activer CORS pour cette réponse
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Décoder le corps de la requête
	var loginReq database.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		log.Printf("Erreur de décodage du corps de la requête: %v", err)
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Authentifier l'utilisateur
	log.Printf("Tentative d'authentification pour: %s", loginReq.Identifier)
	user, err := database.AuthenticateUser(loginReq.Identifier, loginReq.Password)
	if err != nil {
		log.Printf("Échec d'authentification pour %s: %v", loginReq.Identifier, err)
		http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		return
	}
	log.Printf("Authentification réussie pour: %s", user.Username)

	// Créer une session pour l'utilisateur
	log.Printf("Création d'une session pour l'utilisateur: %d", user.ID)
	session, err := database.CreateSession(user.ID)
	if err != nil {
		log.Printf("Erreur lors de la création de la session: %v", err)
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}
	log.Printf("Session créée avec succès: %s", session.ID)

	// Définir le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: false, // Permettre l'accès via JavaScript pour diagnostic
		Path:     "/",
		SameSite: http.SameSiteNoneMode, // Permettre l'accès depuis n'importe quelle origine
		Secure:   false,                 // Ne pas exiger HTTPS en développement
	})

	// Mettre à jour le statut en ligne
	err = database.UpdateUserOnlineStatus(user.ID, true)
	if err != nil {
		// Log l'erreur mais continuer
		log.Printf("Erreur lors de la mise à jour du statut en ligne: %v", err)
	}

	// Nettoyer le mot de passe avant de retourner l'utilisateur
	user.Password = ""

	// Retourner l'utilisateur connecté et l'ID de session
	response := struct {
		User      *database.User `json:"user"`
		SessionID string         `json:"sessionId"`
	}{
		User:      user,
		SessionID: session.ID,
	}

	// Définir le type de contenu avant d'écrire quoi que ce soit
	w.Header().Set("Content-Type", "application/json")

	// Sérialiser manuellement pour éviter les problèmes
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de la réponse: %v", err)
		http.Error(w, "Erreur lors de la sérialisation de la réponse", http.StatusInternalServerError)
		return
	}

	// Écrire directement dans le ResponseWriter
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Printf("Erreur lors de l'écriture de la réponse: %v", err)
	}
}

// LogoutHandler gère la déconnexion d'un utilisateur
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer le cookie de session
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte
	userID, ok := middleware.GetUserID(r)
	if ok {
		// Mettre à jour le statut en ligne
		err = database.UpdateUserOnlineStatus(userID, false)
		if err != nil {
			// Log l'erreur mais continuer
			println("Erreur lors de la mise à jour du statut en ligne:", err.Error())
		}
	}

	// Supprimer la session
	err = database.DeleteSession(cookie.Value)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression de la session", http.StatusInternalServerError)
		return
	}

	// Supprimer le cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	// Retourner un succès
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Déconnecté avec succès"}`))
}

// GetCurrentUserHandler récupère l'utilisateur actuellement connecté
func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Activer CORS pour cette réponse
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	var userID int

	// Vérifier d'abord le cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// Vérifier l'en-tête Authorization
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			session, err := database.GetSessionByID(token)
			if err == nil {
				// Utiliser l'ID utilisateur de la session
				userID = session.UserID
				log.Printf("Authentification par token Bearer réussie pour l'utilisateur ID=%d", userID)
			} else {
				log.Printf("Token Bearer invalide: %v", err)
				http.Error(w, "Non authentifié", http.StatusUnauthorized)
				return
			}
		} else {
			log.Printf("Aucun cookie ou token Bearer trouvé")
			http.Error(w, "Non authentifié", http.StatusUnauthorized)
			return
		}
	} else {
		// Utiliser le cookie comme avant
		session, err := database.GetSessionByID(cookie.Value)
		if err != nil {
			log.Printf("Cookie de session invalide: %v", err)
			http.Error(w, "Session invalide", http.StatusUnauthorized)
			return
		}
		userID = session.UserID
		log.Printf("Authentification par cookie réussie pour l'utilisateur ID=%d", userID)
	}

	// Récupérer l'utilisateur
	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("Erreur lors de la récupération de l'utilisateur: %v", err)
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}
	log.Printf("Utilisateur récupéré avec succès: %s", user.Username)

	// Nettoyer le mot de passe avant de retourner l'utilisateur
	user.Password = ""

	// Définir le type de contenu
	w.Header().Set("Content-Type", "application/json")

	// Sérialiser manuellement pour éviter les problèmes
	responseBytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de la réponse: %v", err)
		http.Error(w, "Erreur lors de la sérialisation de la réponse", http.StatusInternalServerError)
		return
	}

	// Écrire directement dans le ResponseWriter
	_, err = w.Write(responseBytes)
	if err != nil {
		log.Printf("Erreur lors de l'écriture de la réponse: %v", err)
	}
}

// GetOnlineUsersHandler récupère la liste des utilisateurs en ligne
func GetOnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte
	_, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer les utilisateurs en ligne
	users, err := database.GetOnlineUsers()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des utilisateurs en ligne", http.StatusInternalServerError)
		return
	}

	// Retourner la liste des utilisateurs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
