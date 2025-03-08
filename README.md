# Forum en Temps Réel (SPA)

Une application de forum en temps réel développée comme une Single Page Application (SPA) avec Go pour le backend et JavaScript vanilla pour le frontend, utilisant WebSockets pour les communications en temps réel.

## Fonctionnalités

- Architecture SPA (Single Page Application)
- Authentification utilisateur (inscription, connexion, déconnexion)
- Création et consultation de publications
- Commentaires sur les publications
- Messagerie privée en temps réel
- Liste des utilisateurs en ligne
- Indicateur de frappe en temps réel
- Interface utilisateur réactive

## Prérequis

- Go 1.18 ou plus récent
- SQLite3
- Packages Go requis (installés automatiquement via `go mod`)

## Installation

1. Clonez le dépôt :
```bash
git clone [url-du-repo]
cd [nom-du-repo]
```

2. Téléchargez les dépendances Go :
```bash
go mod download
```

3. Compilez l'application :
```bash
go build -o forum
```

## Démarrage

1. Lancez l'application :
```bash
./forum
```

2. Ouvrez votre navigateur et accédez à :
```
http://localhost:8080
```

## Structure du projet

```
.
├── main.go                 # Point d'entrée principal
├── database                # Gestion de la base de données
│   ├── database.go         # Initialisation de la BD
│   ├── models.go           # Modèles de données
│   └── queries.go          # Requêtes SQL
├── handlers                # Gestionnaires HTTP
│   ├── auth.go             # Authentification
│   ├── posts.go            # Publications et commentaires
│   ├── messages.go         # Messages privés
│   └── websocket.go        # WebSockets
├── middleware              # Middleware
│   └── auth.go             # Authentification
├── routes                  # Configuration des routes
│   └── routes.go
├── static                  # Fichiers statiques
│   ├── css
│   │   └── styles.css      # Styles CSS
│   ├── js
│   │   ├── app.js          # Script principal
│   │   ├── auth.js         # Authentification côté client
│   │   ├── posts.js        # Publications côté client
│   │   ├── messages.js     # Messages côté client
│   │   ├── websocket.js    # WebSockets côté client
│   │   └── ui.js           # Interface utilisateur
│   └── index.html          # Page HTML unique (SPA)
└── schema.sql              # Schéma de la base de données
```

## Notes techniques

- L'application utilise SQLite comme base de données, le fichier `forum.db` est créé automatiquement au premier démarrage.
- La communication en temps réel est assurée par des WebSockets (Gorilla WebSocket).
- L'authentification utilise des sessions avec des cookies.
- Le frontend est développé en JavaScript vanilla sans framework.
- La structure SPA permet une navigation fluide sans rechargement de page.

## Développement

Pour contribuer au projet ou le modifier :

1. Les modifications du backend (Go) nécessitent une recompilation :
```bash
go build -o forum
```

2. Les modifications du frontend (HTML, CSS, JavaScript) sont prises en compte immédiatement en rafraîchissant le navigateur.

3. Les modifications du schéma de la base de données nécessitent de supprimer le fichier `forum.db` pour qu'il soit recréé au prochain démarrage :
```bash
rm forum.db
```
