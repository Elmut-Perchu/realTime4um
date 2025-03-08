// fichier: handlers/messages.go
package handlers

import (
	"encoding/json"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/middleware"
	"strconv"
	"strings"
)

// SendPrivateMessageHandler gère l'envoi d'un message privé
func SendPrivateMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte
	senderID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Décoder le corps de la requête
	var message database.PrivateMessage
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if message.Content == "" || message.ReceiverID <= 0 {
		http.Error(w, "Données incomplètes", http.StatusBadRequest)
		return
	}

	// Vérifier que l'expéditeur n'est pas le destinataire
	if senderID == message.ReceiverID {
		http.Error(w, "Impossible d'envoyer un message à soi-même", http.StatusBadRequest)
		return
	}

	// Vérifier que le destinataire existe
	_, err = database.GetUserByID(message.ReceiverID)
	if err != nil {
		http.Error(w, "Destinataire non trouvé", http.StatusNotFound)
		return
	}

	// Définir l'ID de l'expéditeur
	message.SenderID = senderID

	// Créer le message
	messageID, err := database.CreatePrivateMessage(&message)
	if err != nil {
		http.Error(w, "Erreur lors de l'envoi du message", http.StatusInternalServerError)
		return
	}

	// Récupérer le message créé pour avoir toutes les informations (noms, dates...)
	var createdMessage *database.PrivateMessage
	messages, err := database.GetPrivateMessagesByUsers(senderID, message.ReceiverID, 1, 0)
	if err == nil && len(messages) > 0 {
		createdMessage = messages[0]
	} else {
		// Fallback si on ne peut pas récupérer le message complet
		createdMessage = &database.PrivateMessage{
			ID:         messageID,
			SenderID:   senderID,
			ReceiverID: message.ReceiverID,
			Content:    message.Content,
		}
	}

	// Retourner le message créé
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMessage)
}

// GetPrivateMessagesHandler récupère les messages entre l'utilisateur courant et un autre utilisateur
func GetPrivateMessagesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extraire l'ID de l'autre utilisateur de l'URL
	// Format attendu: /api/messages/{user_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	otherUserIDStr := pathParts[3]
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'autre utilisateur existe
	_, err = database.GetUserByID(otherUserID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusNotFound)
		return
	}

	// Récupérer les paramètres de pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Valeur par défaut
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Valeur par défaut
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// Récupérer les messages
	messages, err := database.GetPrivateMessagesByUsers(userID, otherUserID, limit, offset)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages", http.StatusInternalServerError)
		return
	}

	// Marquer les messages comme lus (ceux envoyés par l'autre utilisateur)
	err = database.MarkMessagesAsRead(otherUserID, userID)
	if err != nil {
		// Log l'erreur mais continuer
		println("Erreur lors du marquage des messages comme lus:", err.Error())
	}

	// Retourner les messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// UpdateTypingStatusHandler met à jour le statut de frappe d'un utilisateur
func UpdateTypingStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'ID utilisateur depuis le contexte
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Décoder le corps de la requête
	var typingData struct {
		TargetUserID int  `json:"targetUserId"`
		IsTyping     bool `json:"isTyping"`
	}
	err := json.NewDecoder(r.Body).Decode(&typingData)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if typingData.TargetUserID <= 0 {
		http.Error(w, "ID utilisateur cible invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'utilisateur cible existe
	_, err = database.GetUserByID(typingData.TargetUserID)
	if err != nil {
		http.Error(w, "Utilisateur cible non trouvé", http.StatusNotFound)
		return
	}

	// Mettre à jour le statut de frappe
	err = database.UpdateTypingStatus(userID, typingData.TargetUserID, typingData.IsTyping)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour du statut de frappe", http.StatusInternalServerError)
		return
	}

	// Récupérer le statut complet (avec nom d'utilisateur, etc.)
	indicator, err := database.GetTypingStatus(userID, typingData.TargetUserID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du statut de frappe", http.StatusInternalServerError)
		return
	}

	// Retourner le statut
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicator)
}

// GetTypingStatusHandler récupère le statut de frappe entre l'utilisateur courant et un autre utilisateur
func GetTypingStatusHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extraire l'ID de l'autre utilisateur de l'URL
	// Format attendu: /api/typing/{user_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	otherUserIDStr := pathParts[3]
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	// Récupérer le statut de frappe de l'autre utilisateur vers l'utilisateur courant
	indicator, err := database.GetTypingStatus(otherUserID, userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du statut de frappe", http.StatusInternalServerError)
		return
	}

	// Retourner le statut
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicator)
}
