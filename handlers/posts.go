// fichier: handlers/posts.go
package handlers

import (
	"encoding/json"
	"net/http"
	"realtimeforum/database"
	"realtimeforum/middleware"
	"strconv"
	"strings"
)

// CreatePostHandler gère la création d'une nouvelle publication
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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
	var post database.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if post.Title == "" || post.Content == "" || post.CategoryID <= 0 {
		http.Error(w, "Données incomplètes", http.StatusBadRequest)
		return
	}

	// Définir l'ID utilisateur
	post.UserID = userID

	// Créer la publication
	postID, err := database.CreatePost(&post)
	if err != nil {
		http.Error(w, "Erreur lors de la création de la publication", http.StatusInternalServerError)
		return
	}

	// Récupérer la publication créée
	createdPost, err := database.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de la publication", http.StatusInternalServerError)
		return
	}

	// Retourner la publication créée
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPost)
}

// GetPostsHandler récupère toutes les publications ou filtre par catégorie
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier s'il y a un filtre par catégorie
	categoryIDStr := r.URL.Query().Get("category")
	var posts []*database.Post
	var err error

	if categoryIDStr != "" {
		// Convertir l'ID de catégorie en entier
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "ID de catégorie invalide", http.StatusBadRequest)
			return
		}

		// Récupérer les publications par catégorie
		posts, err = database.GetPostsByCategory(categoryID)
	} else {
		// Récupérer toutes les publications
		posts, err = database.GetAllPosts()
	}

	if err != nil {
		http.Error(w, "Erreur lors de la récupération des publications", http.StatusInternalServerError)
		return
	}

	// Retourner les publications
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// GetPostHandler récupère une publication par son ID
func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID de la publication de l'URL
	// Format attendu: /api/posts/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	postIDStr := pathParts[3]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Récupérer la publication
	post, err := database.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Publication non trouvée", http.StatusNotFound)
		return
	}

	// Retourner la publication
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// CreateCommentHandler gère la création d'un nouveau commentaire
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extraire l'ID de la publication de l'URL
	// Format attendu: /api/posts/{id}/comments
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	postIDStr := pathParts[3]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que la publication existe
	_, err = database.GetPostByID(postID)
	if err != nil {
		http.Error(w, "Publication non trouvée", http.StatusNotFound)
		return
	}

	// Décoder le corps de la requête
	var comment database.Comment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Validation basique
	if comment.Content == "" {
		http.Error(w, "Contenu vide", http.StatusBadRequest)
		return
	}

	// Définir l'ID utilisateur et l'ID de la publication
	comment.UserID = userID
	comment.PostID = postID

	// Créer le commentaire
	commentID, err := database.CreateComment(&comment)
	if err != nil {
		http.Error(w, "Erreur lors de la création du commentaire", http.StatusInternalServerError)
		return
	}

	// Récupérer tous les commentaires de la publication
	comments, err := database.GetCommentsByPostID(postID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
		return
	}

	// Trouver le commentaire créé
	var createdComment *database.Comment
	for _, c := range comments {
		if c.ID == commentID {
			createdComment = c
			break
		}
	}

	if createdComment == nil {
		http.Error(w, "Commentaire créé non trouvé", http.StatusInternalServerError)
		return
	}

	// Retourner le commentaire créé
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdComment)
}

// GetCommentsHandler récupère les commentaires d'une publication
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Extraire l'ID de la publication de l'URL
	// Format attendu: /api/posts/{id}/comments
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "URL invalide", http.StatusBadRequest)
		return
	}

	postIDStr := pathParts[3]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Récupérer les commentaires
	comments, err := database.GetCommentsByPostID(postID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
		return
	}

	// Retourner les commentaires
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// GetCategoriesHandler récupère toutes les catégories
func GetCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode
	if r.Method != http.MethodGet {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer les catégories
	categories, err := database.GetAllCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	// Retourner les catégories
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
