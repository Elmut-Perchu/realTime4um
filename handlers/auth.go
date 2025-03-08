// fichier: handlers/auth.go
package handlers

import (
	"encoding/json"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/middleware"
	"time"
)

// RegisterHandler gère l'inscription d'un nouvel utilisateur
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Décoder le corps de la requête
	var userDTO database.UserDTO
	err := json.NewDecoder(r.Body).Decode(&userDTO)
	if err != nil {
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
	userID, err := database.CreateUser(userDTO)
	if err != nil {
		http.Error(w, "Erreur lors de la création de l'utilisateur: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Créer une session pour l'utilisateur
	session, err := database.CreateSession(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}

	// Définir le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	// Récupérer l'utilisateur créé
	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}

	// Nettoyer le mot de passe avant de retourner l'utilisateur
	user.Password = ""

	// Retourner l'utilisateur créé
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// LoginHandler gère la connexion d'un utilisateur
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Décoder le corps de la requête
	var loginReq database.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Authentifier l'utilisateur
	user, err := database.AuthenticateUser(loginReq.Identifier, loginReq.Password)
	if err != nil {
		http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		return
	}

	// Créer une session pour l'utilisateur
	session, err := database.CreateSession(user.ID)
	if err != nil {
		http.Error(w, "Erreur lors de la création de la session", http.StatusInternalServerError)
		return
	}

	// Définir le cookie de session
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	// Mettre à jour le statut en ligne
	err = database.UpdateUserOnlineStatus(user.ID, true)
	if err != nil {
		// Log l'erreur mais continuer
		println("Erreur lors de la mise à jour du statut en ligne:", err.Error())
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	// Récupérer l'ID utilisateur depuis le contexte
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Récupérer l'utilisateur
	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de l'utilisateur", http.StatusInternalServerError)
		return
	}

	// Nettoyer le mot de passe avant de retourner l'utilisateur
	user.Password = ""

	// Retourner l'utilisateur
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
