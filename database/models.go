// fichier: database/models.go
package database

import "time"

// User représente un utilisateur du forum
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Ne pas exposer le mot de passe
	CreatedAt time.Time `json:"createdAt"`
	LastLogin time.Time `json:"lastLogin,omitempty"`
	Online    bool      `json:"online"`
}

// UserDTO est utilisé pour l'inscription et la connexion
type UserDTO struct {
	Username  string `json:"username"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// LoginRequest représente les données de connexion
type LoginRequest struct {
	Identifier string `json:"identifier"` // Email ou username
	Password   string `json:"password"`
}

// Session représente une session utilisateur
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Category représente une catégorie de publication
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Post représente une publication
type Post struct {
	ID         int       `json:"id"`
	UserID     int       `json:"userId"`
	Username   string    `json:"username,omitempty"` // Pour l'affichage
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CategoryID int       `json:"categoryId"`
	Category   string    `json:"category,omitempty"` // Pour l'affichage
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Comment représente un commentaire sur une publication
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"postId"`
	UserID    int       `json:"userId"`
	Username  string    `json:"username,omitempty"` // Pour l'affichage
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// PrivateMessage représente un message privé entre utilisateurs
type PrivateMessage struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"senderId"`
	ReceiverID int       `json:"receiverId"`
	Sender     string    `json:"sender,omitempty"`     // Pour l'affichage
	Receiver   string    `json:"receiver,omitempty"`   // Pour l'affichage
	Content    string    `json:"content"`
	Read       bool      `json:"read"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TypingIndicator représente un indicateur de frappe
type TypingIndicator struct {
	UserID       int       `json:"userId"`
	Username     string    `json:"username"`
	TargetUserID int       `json:"targetUserId"`
	IsTyping     bool      `json:"isTyping"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// WebSocketMessage représente un message envoyé via WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
