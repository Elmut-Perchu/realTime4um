// fichier: database/queries.go
package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ==================================
// User Operations
// ==================================

// CreateUser crée un nouvel utilisateur dans la base de données
func CreateUser(user UserDTO) (int, error) {
	// Hacher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Insérer le nouvel utilisateur
	result, err := DB.Exec(
		"INSERT INTO users (username, age, gender, first_name, last_name, email, password) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.Username, user.Age, user.Gender, user.FirstName, user.LastName, user.Email, string(hashedPassword),
	)
	if err != nil {
		return 0, err
	}

	// Obtenir l'ID généré
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetUserByID récupère un utilisateur par son ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, username, age, gender, first_name, last_name, email, password, created_at, last_login, online FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.LastLogin, &user.Online)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("utilisateur non trouvé")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByEmail récupère un utilisateur par son email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, username, age, gender, first_name, last_name, email, password, created_at, last_login, online FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.LastLogin, &user.Online)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("utilisateur non trouvé")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername récupère un utilisateur par son nom d'utilisateur
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, username, age, gender, first_name, last_name, email, password, created_at, last_login, online FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.LastLogin, &user.Online)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("utilisateur non trouvé")
		}
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authentifie un utilisateur par identifiant (email ou username) et mot de passe
func AuthenticateUser(identifier, password string) (*User, error) {
	var user *User
	var err error

	// Vérifier si l'identifiant est un email ou un nom d'utilisateur
	if strings.Contains(identifier, "@") {
		user, err = GetUserByEmail(identifier)
	} else {
		user, err = GetUserByUsername(identifier)
	}

	if err != nil {
		return nil, err
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("mot de passe incorrect")
	}

	// Mettre à jour last_login et online
	_, err = DB.Exec("UPDATE users SET last_login = ?, online = TRUE WHERE id = ?", time.Now(), user.ID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserOnlineStatus met à jour le statut en ligne d'un utilisateur
func UpdateUserOnlineStatus(userID int, online bool) error {
	_, err := DB.Exec("UPDATE users SET online = ? WHERE id = ?", online, userID)
	return err
}

// GetOnlineUsers récupère tous les utilisateurs en ligne triés par dernier message
func GetOnlineUsers() ([]*User, error) {
	rows, err := DB.Query(`
		SELECT u.id, u.username, u.age, u.gender, u.first_name, u.last_name, u.email, u.created_at, u.last_login, u.online
		FROM users u
		LEFT JOIN (
			SELECT sender_id, MAX(created_at) as last_msg
			FROM private_messages
			GROUP BY sender_id
		) pm ON u.id = pm.sender_id
		WHERE u.online = TRUE
		ORDER BY pm.last_msg DESC NULLS LAST, u.username ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt, &user.LastLogin, &user.Online)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ==================================
// Session Operations
// ==================================

// CreateSession crée une nouvelle session pour un utilisateur
func CreateSession(userID int) (*Session, error) {
	// Générer un ID de session unique
	sessionID := uuid.NewString()

	// Définir l'expiration (24 heures)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Insérer la session dans la base de données
	_, err := DB.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

// GetSessionByID récupère une session par son ID
func GetSessionByID(sessionID string) (*Session, error) {
	session := &Session{}
	err := DB.QueryRow(
		"SELECT id, user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session non trouvée")
		}
		return nil, err
	}

	// Vérifier si la session est expirée
	if time.Now().After(session.ExpiresAt) {
		// Supprimer la session expirée
		_, _ = DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		return nil, errors.New("session expirée")
	}

	return session, nil
}

// DeleteSession supprime une session
func DeleteSession(sessionID string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// ==================================
// Post Operations
// ==================================

// CreatePost crée une nouvelle publication
func CreatePost(post *Post) (int, error) {
	result, err := DB.Exec(
		"INSERT INTO posts (user_id, title, content, category_id) VALUES (?, ?, ?, ?)",
		post.UserID, post.Title, post.Content, post.CategoryID,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetPostByID récupère une publication par son ID
func GetPostByID(postID int) (*Post, error) {
	post := &Post{}
	err := DB.QueryRow(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.category_id, c.name, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content,
		&post.CategoryID, &post.Category, &post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("publication non trouvée")
		}
		return nil, err
	}

	return post, nil
}

// GetAllPosts récupère toutes les publications
func GetAllPosts() ([]*Post, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.category_id, c.name, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content,
			&post.CategoryID, &post.Category, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostsByCategory récupère les publications par catégorie
func GetPostsByCategory(categoryID int) ([]*Post, error) {
	rows, err := DB.Query(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.category_id, c.name, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*Post, 0)
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content,
			&post.CategoryID, &post.Category, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetAllCategories récupère toutes les catégories
func GetAllCategories() ([]*Category, error) {
	rows, err := DB.Query("SELECT id, name, description FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*Category, 0)
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// ==================================
// Comment Operations
// ==================================

// CreateComment crée un nouveau commentaire
func CreateComment(comment *Comment) (int, error) {
	result, err := DB.Exec(
		"INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
		comment.PostID, comment.UserID, comment.Content,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetCommentsByPostID récupère les commentaires d'une publication
func GetCommentsByPostID(postID int) ([]*Comment, error) {
	rows, err := DB.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*Comment, 0)
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Username,
			&comment.Content, &comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// ==================================
// Private Message Operations
// ==================================

// CreatePrivateMessage crée un nouveau message privé
func CreatePrivateMessage(message *PrivateMessage) (int, error) {
	result, err := DB.Exec(
		"INSERT INTO private_messages (sender_id, receiver_id, content) VALUES (?, ?, ?)",
		message.SenderID, message.ReceiverID, message.Content,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetPrivateMessagesByUsers récupère les messages entre deux utilisateurs avec limite et pagination
func GetPrivateMessagesByUsers(userID1, userID2 int, limit, offset int) ([]*PrivateMessage, error) {
	rows, err := DB.Query(`
		SELECT pm.id, pm.sender_id, pm.receiver_id, s.username, r.username, pm.content, pm.read, pm.created_at
		FROM private_messages pm
		JOIN users s ON pm.sender_id = s.id
		JOIN users r ON pm.receiver_id = r.id
		WHERE (pm.sender_id = ? AND pm.receiver_id = ?) OR (pm.sender_id = ? AND pm.receiver_id = ?)
		ORDER BY pm.created_at DESC
		LIMIT ? OFFSET ?
	`, userID1, userID2, userID2, userID1, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*PrivateMessage, 0)
	for rows.Next() {
		message := &PrivateMessage{}
		err := rows.Scan(
			&message.ID, &message.SenderID, &message.ReceiverID, &message.Sender, &message.Receiver,
			&message.Content, &message.Read, &message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Inverser l'ordre pour obtenir les plus anciens en premier
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// MarkMessagesAsRead marque les messages comme lus
func MarkMessagesAsRead(senderID, receiverID int) error {
	_, err := DB.Exec(
		"UPDATE private_messages SET read = TRUE WHERE sender_id = ? AND receiver_id = ? AND read = FALSE",
		senderID, receiverID,
	)
	return err
}

// ==================================
// Typing Indicator Operations
// ==================================

// UpdateTypingStatus met à jour le statut de frappe d'un utilisateur
func UpdateTypingStatus(userID, targetUserID int, isTyping bool) error {
	// Vérifier si un enregistrement existe déjà
	var count int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM typing_indicators WHERE user_id = ? AND target_user_id = ?",
		userID, targetUserID,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Mettre à jour l'enregistrement existant
		_, err = DB.Exec(
			"UPDATE typing_indicators SET is_typing = ?, updated_at = ? WHERE user_id = ? AND target_user_id = ?",
			isTyping, time.Now(), userID, targetUserID,
		)
	} else {
		// Créer un nouvel enregistrement
		_, err = DB.Exec(
			"INSERT INTO typing_indicators (user_id, target_user_id, is_typing, updated_at) VALUES (?, ?, ?, ?)",
			userID, targetUserID, isTyping, time.Now(),
		)
	}

	return err
}

// GetTypingStatus récupère le statut de frappe entre deux utilisateurs
func GetTypingStatus(userID, targetUserID int) (*TypingIndicator, error) {
	indicator := &TypingIndicator{}
	err := DB.QueryRow(`
		SELECT ti.user_id, u.username, ti.target_user_id, ti.is_typing, ti.updated_at
		FROM typing_indicators ti
		JOIN users u ON ti.user_id = u.id
		WHERE ti.user_id = ? AND ti.target_user_id = ?
	`, userID, targetUserID).Scan(
		&indicator.UserID, &indicator.Username, &indicator.TargetUserID,
		&indicator.IsTyping, &indicator.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Pas d'indicateur trouvé, retourner un indicateur par défaut
			return &TypingIndicator{
				UserID:       userID,
				TargetUserID: targetUserID,
				IsTyping:     false,
				UpdatedAt:    time.Now(),
			}, nil
		}
		return nil, err
	}

	return indicator, nil
}
