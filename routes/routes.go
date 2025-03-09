package routes

import (
	"net/http"
	"realtimeforum/handlers"
	"realtimeforum/middleware"
	"strings"
)

// apiHandler est un gestionnaire personnalisé qui utilise un switch pour router les requêtes
type apiHandler struct{}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Activer CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Traiter les requêtes OPTIONS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Router les requêtes en fonction du chemin
	switch {
	// Routes d'authentification
	case r.URL.Path == "/api/register":
		handlers.RegisterHandler(w, r)
	case r.URL.Path == "/api/login":
		handlers.LoginHandler(w, r)
	case r.URL.Path == "/api/logout":
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.LogoutHandler))
		authHandler.ServeHTTP(w, r)
	case r.URL.Path == "/api/me":
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.GetCurrentUserHandler))
		authHandler.ServeHTTP(w, r)
	case r.URL.Path == "/api/users/online":
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.GetOnlineUsersHandler))
		authHandler.ServeHTTP(w, r)

	// Routes des publications et commentaires
	case r.URL.Path == "/api/posts" && r.Method == http.MethodGet:
		optionalAuthHandler := middleware.OptionalAuthMiddleware(http.HandlerFunc(handlers.GetPostsHandler))
		optionalAuthHandler.ServeHTTP(w, r)
	case r.URL.Path == "/api/posts" && r.Method == http.MethodPost:
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler))
		authHandler.ServeHTTP(w, r)
	case r.URL.Path == "/api/categories":
		handlers.GetCategoriesHandler(w, r)
	case len(r.URL.Path) > 10 && r.URL.Path[:10] == "/api/posts/" && r.Method == http.MethodGet:
		if len(r.URL.Path) > 19 && r.URL.Path[len(r.URL.Path)-9:] == "/comments" {
			// Route pour les commentaires d'une publication
			handlers.GetCommentsHandler(w, r)
		} else {
			// Route pour une publication spécifique
			handlers.GetPostHandler(w, r)
		}
	case len(r.URL.Path) > 19 && r.URL.Path[:19] == "/api/posts/" && r.URL.Path[len(r.URL.Path)-9:] == "/comments" && r.Method == http.MethodPost:
		// Route pour créer un commentaire
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.CreateCommentHandler))
		authHandler.ServeHTTP(w, r)

	// Routes des messages privés
	case r.URL.Path == "/api/messages" && r.Method == http.MethodPost:
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.SendPrivateMessageHandler))
		authHandler.ServeHTTP(w, r)
	case len(r.URL.Path) > 14 && r.URL.Path[:14] == "/api/messages/" && r.Method == http.MethodGet:
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.GetPrivateMessagesHandler))
		authHandler.ServeHTTP(w, r)

	// Routes pour l'indicateur de frappe
	case r.URL.Path == "/api/typing" && r.Method == http.MethodPost:
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdateTypingStatusHandler))
		authHandler.ServeHTTP(w, r)
	case len(r.URL.Path) > 12 && r.URL.Path[:12] == "/api/typing/" && r.Method == http.MethodGet:
		authHandler := middleware.AuthMiddleware(http.HandlerFunc(handlers.GetTypingStatusHandler))
		authHandler.ServeHTTP(w, r)

	// Route par défaut
	default:
		http.NotFound(w, r)
	}
}

// SetupRoutes configure toutes les routes de l'application
func SetupRoutes() http.Handler {
	// Créer un nouveau multiplexeur
	mux := http.NewServeMux()

	// Ajouter un gestionnaire CORS global
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Configurer les en-têtes CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Répondre immédiatement aux requêtes OPTIONS
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Appeler le gestionnaire suivant
			h.ServeHTTP(w, r)
		})
	}

	// Appliquer le middleware CORS à toutes les routes
	handler := corsHandler(apiHandler{})
	mux.Handle("/api/", handler)

	// Ajouter le gestionnaire WebSocket avec CORS
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Appliquer CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Traiter la requête WebSocket
		handlers.WebSocketHandler(w, r)
	})

	// Servir les fichiers statiques avec gestion du SPA
	fileServer := http.FileServer(http.Dir("static"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Appliquer CORS pour les fichiers statiques aussi
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		path := r.URL.Path

		// Si c'est un fichier JavaScript ou CSS, le servir directement
		if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Vérifier si le fichier existe
		if _, err := http.Dir("static").Open(path); err != nil && path != "/" {
			// Si le fichier n'existe pas et que ce n'est pas la racine,
			// servir index.html pour permettre au frontend de gérer les routes
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}
