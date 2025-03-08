// fichier: handlers/websocket.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/middleware"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	// Upgrader pour convertir une connexion HTTP en WebSocket
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Permettre toutes les origines (à adapter en production)
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Clients stocke toutes les connexions WebSocket actives
	// La clé est l'ID de l'utilisateur
	clients = make(map[int]*Client)

	// Mutex pour protéger l'accès à la map clients
	clientsMutex = sync.RWMutex{}
)

// Client représente un client WebSocket connecté
type Client struct {
	UserID int
	Conn   *websocket.Conn
	Send   chan []byte
}

// Message représente un message WebSocket
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketHandler gère les connexions WebSocket
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Authentifier l'utilisateur
	userID, ok := middleware.WSAuthMiddleware(w, r)
	if !ok {
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	// Mettre à jour le statut en ligne
	err := database.UpdateUserOnlineStatus(userID, true)
	if err != nil {
		log.Printf("Erreur lors de la mise à jour du statut en ligne: %v", err)
	}

	// Mettre à niveau la connexion HTTP vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erreur lors de l'upgrade de la connexion: %v", err)
		return
	}

	// Créer un nouveau client
	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	// Enregistrer le client
	clientsMutex.Lock()
	// Fermer la connexion précédente si elle existe
	if existingClient, ok := clients[userID]; ok {
		close(existingClient.Send)
		existingClient.Conn.Close()
	}
	clients[userID] = client
	clientsMutex.Unlock()

	// Envoyer la liste des utilisateurs en ligne à tous les clients
	broadcastOnlineUsers()

	// Démarrer les goroutines pour la lecture et l'écriture
	go client.readPump()
	go client.writePump()
}

// readPump pompe les messages du client WebSocket vers le hub
func (c *Client) readPump() {
	defer func() {
		// Fermer la connexion quand on sort de la fonction
		c.Conn.Close()

		// Supprimer le client de la map des clients
		clientsMutex.Lock()
		delete(clients, c.UserID)
		clientsMutex.Unlock()

		// Mettre à jour le statut en ligne
		err := database.UpdateUserOnlineStatus(c.UserID, false)
		if err != nil {
			log.Printf("Erreur lors de la mise à jour du statut en ligne: %v", err)
		}

		// Diffuser la mise à jour des utilisateurs en ligne
		broadcastOnlineUsers()

		// Fermer le canal d'envoi
		close(c.Send)
	}()

	// Configurer le WebSocket
	c.Conn.SetReadLimit(4096) // 4KB max par message

	// Lire les messages
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erreur: %v", err)
			}
			break
		}

		// Traiter le message
		processMessage(c.UserID, message)
	}
}

// writePump pompe les messages du hub vers le client WebSocket
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Send
		if !ok {
			// Le canal a été fermé
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Erreur lors de l'envoi du message: %v", err)
			return
		}
	}
}

// processMessage traite un message reçu d'un client
func processMessage(senderID int, rawMessage []byte) {
	// Décoder le message
	var message Message
	err := json.Unmarshal(rawMessage, &message)
	if err != nil {
		log.Printf("Erreur lors du décodage du message: %v", err)
		return
	}

	// Traiter en fonction du type de message
	switch message.Type {
	case "private_message":
		// Traiter l'envoi d'un message privé
		handlePrivateMessage(senderID, message.Payload)
	case "typing_indicator":
		// Traiter l'indicateur de frappe
		handleTypingIndicator(senderID, message.Payload)
	case "post_created":
		// Diffuser une nouvelle publication à tous les clients
		broadcastToAll(rawMessage)
	case "comment_created":
		// Diffuser un nouveau commentaire à tous les clients
		broadcastToAll(rawMessage)
	default:
		log.Printf("Type de message inconnu: %s", message.Type)
	}
}

// handlePrivateMessage traite l'envoi d'un message privé
func handlePrivateMessage(senderID int, payload interface{}) {
	// Convertir le payload en message privé
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Erreur lors de la conversion du payload: %v", err)
		return
	}

	var privateMessage database.PrivateMessage
	err = json.Unmarshal(payloadJSON, &privateMessage)
	if err != nil {
		log.Printf("Erreur lors du décodage du message privé: %v", err)
		return
	}

	// Définir l'ID de l'expéditeur
	privateMessage.SenderID = senderID

	// Sauvegarder le message dans la base de données
	_, err = database.CreatePrivateMessage(&privateMessage)
	if err != nil {
		log.Printf("Erreur lors de l'enregistrement du message: %v", err)
		return
	}

	// Récupérer le message complet
	messages, err := database.GetPrivateMessagesByUsers(senderID, privateMessage.ReceiverID, 1, 0)
	if err != nil || len(messages) == 0 {
		log.Printf("Erreur lors de la récupération du message: %v", err)
		return
	}

	// Créer le message à envoyer
	wsMessage := Message{
		Type:    "private_message",
		Payload: messages[0],
	}

	messageJSON, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation du message: %v", err)
		return
	}

	// Envoyer le message au destinataire
	sendToUser(privateMessage.ReceiverID, messageJSON)

	// Envoyer une confirmation à l'expéditeur (pour s'assurer que le message a bien été enregistré)
	sendToUser(senderID, messageJSON)
}

// handleTypingIndicator traite un indicateur de frappe
func handleTypingIndicator(userID int, payload interface{}) {
	// Convertir le payload en indicateur de frappe
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Erreur lors de la conversion du payload: %v", err)
		return
	}

	var typingData struct {
		TargetUserID int  `json:"targetUserId"`
		IsTyping     bool `json:"isTyping"`
	}
	err = json.Unmarshal(payloadJSON, &typingData)
	if err != nil {
		log.Printf("Erreur lors du décodage de l'indicateur de frappe: %v", err)
		return
	}

	// Mettre à jour le statut de frappe
	err = database.UpdateTypingStatus(userID, typingData.TargetUserID, typingData.IsTyping)
	if err != nil {
		log.Printf("Erreur lors de la mise à jour du statut de frappe: %v", err)
		return
	}

	// Récupérer le statut complet
	indicator, err := database.GetTypingStatus(userID, typingData.TargetUserID)
	if err != nil {
		log.Printf("Erreur lors de la récupération du statut de frappe: %v", err)
		return
	}

	// Créer le message à envoyer
	wsMessage := Message{
		Type:    "typing_indicator",
		Payload: indicator,
	}

	messageJSON, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation du message: %v", err)
		return
	}

	// Envoyer le message au destinataire
	sendToUser(typingData.TargetUserID, messageJSON)
}

// sendToUser envoie un message à un utilisateur spécifique
func sendToUser(userID int, message []byte) {
	clientsMutex.RLock()
	client, ok := clients[userID]
	clientsMutex.RUnlock()

	if ok {
		select {
		case client.Send <- message:
			// Message envoyé avec succès
		default:
			// Le canal est plein ou fermé, supprimer le client
			clientsMutex.Lock()
			delete(clients, userID)
			clientsMutex.Unlock()
			close(client.Send)
		}
	}
}

// broadcastToAll envoie un message à tous les clients connectés
func broadcastToAll(message []byte) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- message:
			// Message envoyé avec succès
		default:
			// Le canal est plein ou fermé, cela sera nettoyé lors du prochain cycle de lecture
		}
	}
}

// broadcastOnlineUsers diffuse la liste des utilisateurs en ligne à tous les clients
func broadcastOnlineUsers() {
	// Récupérer la liste des utilisateurs en ligne
	users, err := database.GetOnlineUsers()
	if err != nil {
		log.Printf("Erreur lors de la récupération des utilisateurs en ligne: %v", err)
		return
	}

	// Créer le message à envoyer
	wsMessage := Message{
		Type:    "online_users",
		Payload: users,
	}

	messageJSON, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation du message: %v", err)
		return
	}

	// Diffuser à tous les clients
	broadcastToAll(messageJSON)
}
