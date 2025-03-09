// fichier: static/js/app.js
import { initAuth, isAuthenticated, getCurrentUser } from './auth.js';
import { initPosts } from './posts.js';
import { initMessages } from './messages.js';
import { initWebSocket } from './websocket.js';
import { initUI, showPage, updateUI } from './ui.js';

// État global de l'application
const state = {
    currentUser: null,
    isAuthenticated: false,
    currentPage: 'home',
    onlineUsers: [],
    categories: [],
    posts: [],
    currentPost: null,
    currentChatUser: null,
    socket: null
};

// Initialisation de l'application
async function initApp() {
    try {
        console.log("Initialisation de l'application...");

        // Initialiser l'interface utilisateur
        console.log("Initialisation de l'UI...");
        initUI(state, navigateTo);

        // Initialiser l'authentification
        console.log("Initialisation de l'authentification...");
        await initAuth(state, updateAppState);

        // Initialiser les publications
        console.log("Initialisation des publications...");
        initPosts(state, updateAppState);

        // Initialiser les messages
        console.log("Initialisation des messages...");
        initMessages(state);

        // Initialiser les WebSockets si l'utilisateur est authentifié
        if (state.isAuthenticated) {
            console.log("Initialisation des WebSockets...");
            const socket = initWebSocket(state, handleWebSocketMessage);
            state.socket = socket;
        }

        // Gérer la navigation initiale
        console.log("Gestion de la navigation initiale...");
        handleInitialNavigation();

        console.log("Initialisation terminée avec succès!");
    } catch (error) {
        console.error("Erreur lors de l'initialisation de l'application:", error);
    }
}

// Mise à jour de l'état de l'application
function updateAppState(newState) {
    // Fusionner le nouvel état avec l'état actuel
    Object.assign(state, newState);

    // Mettre à jour l'interface utilisateur
    updateUI(state);

    // Si l'authentification a changé, gérer les WebSockets
    if (newState.isAuthenticated !== undefined) {
        handleAuthenticationChange();
    }
}

// Gestion du changement d'authentification
function handleAuthenticationChange() {
    if (state.isAuthenticated) {
        // L'utilisateur vient de se connecter
        if (!state.socket) {
            const socket = initWebSocket(state, handleWebSocketMessage);
            state.socket = socket;
        }
    } else {
        // L'utilisateur vient de se déconnecter
        if (state.socket) {
            state.socket.close();
            state.socket = null;
        }
    }
}

// Gestion des messages WebSocket
function handleWebSocketMessage(event) {
    const message = JSON.parse(event.data);

    switch (message.type) {
        case 'online_users':
            updateAppState({ onlineUsers: message.payload });
            break;
        case 'private_message':
            handlePrivateMessage(message.payload);
            break;
        case 'typing_indicator':
            handleTypingIndicator(message.payload);
            break;
        case 'post_created':
            handleNewPost(message.payload);
            break;
        case 'comment_created':
            handleNewComment(message.payload);
            break;
        default:
            console.log('Type de message non géré:', message.type);
    }
}

// Gestion des messages privés reçus
function handlePrivateMessage(message) {
    const messagesList = document.getElementById('messages-list');
    const isCurrentChat = state.currentChatUser &&
        (message.senderId === state.currentChatUser.id ||
            message.receiverId === state.currentChatUser.id);

    // Si c'est la conversation courante, ajouter le message
    if (isCurrentChat) {
        const isSent = message.senderId === state.currentUser.id;

        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${isSent ? 'sent' : 'received'}`;

        const content = document.createElement('div');
        content.className = 'message-content';
        content.textContent = message.content;

        const time = document.createElement('div');
        time.className = 'message-time';
        time.textContent = new Date(message.createdAt).toLocaleTimeString();

        messageDiv.appendChild(content);
        messageDiv.appendChild(time);
        messagesList.appendChild(messageDiv);

        // Faire défiler vers le bas
        messagesList.scrollTop = messagesList.scrollHeight;
    }

    // Mettre à jour la liste des utilisateurs avec le dernier message
    updateOnlineUsersList();
}

// Gestion des indicateurs de frappe
function handleTypingIndicator(typingData) {
    const typingIndicator = document.getElementById('typing-indicator');
    const typingUsername = document.getElementById('typing-username');

    // Vérifier si c'est pour la conversation courante
    if (state.currentChatUser && typingData.userId === state.currentChatUser.id) {
        if (typingData.isTyping) {
            // Afficher l'indicateur
            typingUsername.textContent = typingData.username;
            typingIndicator.classList.remove('hidden');
        } else {
            // Masquer l'indicateur
            typingIndicator.classList.add('hidden');
        }
    }
}

// Gestion des nouvelles publications
function handleNewPost(post) {
    // Ajouter la nouvelle publication à la liste
    state.posts = [post, ...state.posts];

    // Mettre à jour l'affichage si nécessaire
    if (state.currentPage === 'home') {
        updatePostsList();
    }
}

// Gestion des nouveaux commentaires
function handleNewComment(comment) {
    // Vérifier si c'est pour la publication courante
    if (state.currentPost && comment.postId === state.currentPost.id) {
        // Ajouter le commentaire à la liste
        const commentsList = document.getElementById('comments-list');

        const commentDiv = document.createElement('div');
        commentDiv.className = 'comment';

        const header = document.createElement('div');
        header.className = 'comment-header';

        const author = document.createElement('div');
        author.className = 'comment-author';
        author.textContent = comment.username;

        const date = document.createElement('div');
        date.className = 'comment-date';
        date.textContent = new Date(comment.createdAt).toLocaleString();

        const content = document.createElement('div');
        content.className = 'comment-content';
        content.textContent = comment.content;

        header.appendChild(author);
        header.appendChild(date);
        commentDiv.appendChild(header);
        commentDiv.appendChild(content);

        commentsList.appendChild(commentDiv);
    }
}

// Navigation vers une page
function navigateTo(page, data) {
    console.log(`Navigation vers ${page}`, data);

    // Mettre à jour l'état
    state.currentPage = page;

    // Si des données sont fournies, les stocker
    if (data) {
        if (page === 'post-detail') {
            state.currentPost = data;
        } else if (page === 'messages' && data.user) {
            state.currentChatUser = data.user;
        }
    }

    // Afficher la page
    showPage(page);

    // Mettre à jour l'URL
    updateUrl(page, data);

    // Charger les données appropriées pour la page
    loadPageData(page, data);
}

// Mise à jour de l'URL
function updateUrl(page, data) {
    let url = '/';

    switch (page) {
        case 'home':
            url = '/';
            break;
        case 'categories':
            url = '/categories';
            break;
        case 'messages':
            if (data && data.user) {
                url = `/messages/${data.user.id}`;
            } else {
                url = '/messages';
            }
            break;
        case 'post-detail':
            if (data) {
                url = `/posts/${data.id}`;
            }
            break;
    }

    history.pushState({ page, data }, '', url);
}

// Chargement des données pour une page
function loadPageData(page, data) {
    switch (page) {
        case 'home':
            fetchPosts();
            break;
        case 'categories':
            fetchCategories();
            break;
        case 'messages':
            if (data && data.user) {
                fetchMessages(data.user.id);
            }
            fetchOnlineUsers();
            break;
        case 'post-detail':
            if (data) {
                fetchComments(data.id);
            }
            break;
    }
}

// Gestion de la navigation initiale
function handleInitialNavigation() {
    // Récupérer la page à partir de l'URL
    const path = window.location.pathname;
    let page = 'home';
    let data = null;

    if (path === '/categories') {
        page = 'categories';
    } else if (path.startsWith('/messages')) {
        page = 'messages';
        const userId = path.split('/')[2];
        if (userId && !isNaN(userId)) {
            fetchUserById(userId).then(user => {
                if (user) {
                    navigateTo('messages', { user });
                }
            });
        }
    } else if (path.startsWith('/posts')) {
        page = 'post-detail';
        const postId = path.split('/')[2];
        if (postId && !isNaN(postId)) {
            fetchPostById(postId).then(post => {
                if (post) {
                    navigateTo('post-detail', post);
                }
            });
        }
    }

    // Naviguer vers la page appropriée
    if (page !== 'post-detail' && page !== 'messages') {
        navigateTo(page, data);
    }
}

// Récupération des publications
async function fetchPosts() {
    try {
        const response = await fetch('/api/posts');
        const posts = await response.json();

        updateAppState({ posts });
        updatePostsList();
    } catch (error) {
        console.error('Erreur lors de la récupération des publications:', error);

        // Afficher un message d'erreur à l'utilisateur
        const postsContainer = document.getElementById('posts-container');
        postsContainer.innerHTML = '<div class="error">Erreur lors du chargement des publications. Veuillez réessayer.</div>';
    }
}

// Récupération des catégories
async function fetchCategories() {
    try {
        const response = await fetch('/api/categories');
        const categories = await response.json();

        updateAppState({ categories });
        updateCategoriesList();
    } catch (error) {
        console.error('Erreur lors de la récupération des catégories:', error);

        // Afficher un message d'erreur à l'utilisateur
        const categoriesContainer = document.getElementById('categories-container');
        categoriesContainer.innerHTML = '<div class="error">Erreur lors du chargement des catégories. Veuillez réessayer.</div>';
    }
}

// Récupération des commentaires d'une publication
async function fetchComments(postId) {
    try {
        const response = await fetch(`/api/posts/${postId}/comments`);
        const comments = await response.json();

        updateCommentsList(comments);
    } catch (error) {
        console.error('Erreur lors de la récupération des commentaires:', error);

        // Afficher un message d'erreur à l'utilisateur
        const commentsList = document.getElementById('comments-list');
        commentsList.innerHTML = '<div class="error">Erreur lors du chargement des commentaires. Veuillez réessayer.</div>';
    }
}

// Récupération des messages avec un utilisateur
async function fetchMessages(userId) {
    try {
        const response = await fetch(`/api/messages/${userId}`);
        const messages = await response.json();

        updateMessagesList(messages);
    } catch (error) {
        console.error('Erreur lors de la récupération des messages:', error);

        // Afficher un message d'erreur à l'utilisateur
        const messagesList = document.getElementById('messages-list');
        messagesList.innerHTML = '<div class="error">Erreur lors du chargement des messages. Veuillez réessayer.</div>';
    }
}

// Récupération des utilisateurs en ligne
async function fetchOnlineUsers() {
    try {
        const response = await fetch('/api/users/online');
        const users = await response.json();

        updateAppState({ onlineUsers: users });
        updateOnlineUsersList();
    } catch (error) {
        console.error('Erreur lors de la récupération des utilisateurs en ligne:', error);

        // Afficher un message d'erreur à l'utilisateur
        const onlineUsers = document.getElementById('online-users');
        onlineUsers.innerHTML = '<li class="error">Erreur lors du chargement des utilisateurs. Veuillez réessayer.</li>';
    }
}

// Récupération d'une publication par ID
async function fetchPostById(postId) {
    try {
        const response = await fetch(`/api/posts/${postId}`);

        if (!response.ok) {
            throw new Error('Publication non trouvée');
        }

        return await response.json();
    } catch (error) {
        console.error('Erreur lors de la récupération de la publication:', error);
        return null;
    }
}

// Récupération d'un utilisateur par ID
async function fetchUserById(userId) {
    try {
        // Cette fonction peut récupérer un utilisateur spécifique,
        // mais comme l'API ne fournit pas d'endpoint pour cela,
        // nous utilisons la liste des utilisateurs en ligne
        const response = await fetch('/api/users/online');
        const users = await response.json();

        const user = users.find(u => u.id === parseInt(userId));

        if (!user) {
            throw new Error('Utilisateur non trouvé');
        }

        return user;
    } catch (error) {
        console.error('Erreur lors de la récupération de l\'utilisateur:', error);
        return null;
    }
}

// Mise à jour de la liste des publications
function updatePostsList() {
    const postsContainer = document.getElementById('posts-container');

    if (state.posts.length === 0) {
        postsContainer.innerHTML = '<div class="empty">Aucune publication disponible.</div>';
        return;
    }

    postsContainer.innerHTML = '';

    state.posts.forEach(post => {
        const postCard = document.createElement('div');
        postCard.className = 'post-card';
        postCard.onclick = () => navigateTo('post-detail', post);

        const header = document.createElement('div');
        header.className = 'post-header';

        const titleDiv = document.createElement('div');

        const title = document.createElement('div');
        title.className = 'post-title';
        title.textContent = post.title;

        const category = document.createElement('div');
        category.className = 'post-category';
        category.textContent = post.category;

        const authorDiv = document.createElement('div');

        const author = document.createElement('div');
        author.className = 'post-author';
        author.textContent = post.username;

        const date = document.createElement('div');
        date.className = 'post-date';
        date.textContent = new Date(post.createdAt).toLocaleString();

        const content = document.createElement('div');
        content.className = 'post-content';
        // Tronquer le contenu s'il est trop long
        content.textContent = post.content.length > 200
            ? post.content.substring(0, 200) + '...'
            : post.content;

        titleDiv.appendChild(title);
        titleDiv.appendChild(category);
        authorDiv.appendChild(author);
        authorDiv.appendChild(date);
        header.appendChild(titleDiv);
        header.appendChild(authorDiv);
        postCard.appendChild(header);
        postCard.appendChild(content);

        postsContainer.appendChild(postCard);
    });
}

// Mise à jour de la liste des catégories
function updateCategoriesList() {
    const categoriesContainer = document.getElementById('categories-container');

    if (state.categories.length === 0) {
        categoriesContainer.innerHTML = '<div class="empty">Aucune catégorie disponible.</div>';
        return;
    }

    categoriesContainer.innerHTML = '';

    state.categories.forEach(category => {
        const categoryCard = document.createElement('div');
        categoryCard.className = 'category-card';
        categoryCard.onclick = () => {
            // Rediriger vers la page d'accueil avec le filtre de catégorie
            window.location.href = `/?category=${category.id}`;
        };

        const name = document.createElement('div');
        name.className = 'category-name';
        name.textContent = category.name;

        const description = document.createElement('div');
        description.className = 'category-description';
        description.textContent = category.description;

        categoryCard.appendChild(name);
        categoryCard.appendChild(description);

        categoriesContainer.appendChild(categoryCard);
    });

    // Mettre également à jour le select dans le formulaire de nouvelle publication
    const postCategorySelect = document.getElementById('post-category');
    postCategorySelect.innerHTML = '';

    state.categories.forEach(category => {
        const option = document.createElement('option');
        option.value = category.id;
        option.textContent = category.name;
        postCategorySelect.appendChild(option);
    });
}

// Mise à jour de la liste des commentaires
function updateCommentsList(comments) {
    const commentsList = document.getElementById('comments-list');

    if (!comments || comments.length === 0) {
        commentsList.innerHTML = '<div class="empty">Aucun commentaire pour cette publication.</div>';
        return;
    }

    commentsList.innerHTML = '';

    comments.forEach(comment => {
        const commentDiv = document.createElement('div');
        commentDiv.className = 'comment';

        const header = document.createElement('div');
        header.className = 'comment-header';

        const author = document.createElement('div');
        author.className = 'comment-author';
        author.textContent = comment.username;

        const date = document.createElement('div');
        date.className = 'comment-date';
        date.textContent = new Date(comment.createdAt).toLocaleString();

        const content = document.createElement('div');
        content.className = 'comment-content';
        content.textContent = comment.content;

        header.appendChild(author);
        header.appendChild(date);
        commentDiv.appendChild(header);
        commentDiv.appendChild(content);

        commentsList.appendChild(commentDiv);
    });
}

// Mise à jour de la liste des messages
function updateMessagesList(messages) {
    const messagesList = document.getElementById('messages-list');

    if (!messages || messages.length === 0) {
        messagesList.innerHTML = '<div class="empty">Aucun message dans cette conversation.</div>';
        return;
    }

    messagesList.innerHTML = '';

    messages.forEach(message => {
        const isSent = message.senderId === state.currentUser.id;

        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${isSent ? 'sent' : 'received'}`;

        const content = document.createElement('div');
        content.className = 'message-content';
        content.textContent = message.content;

        const time = document.createElement('div');
        time.className = 'message-time';
        time.textContent = new Date(message.createdAt).toLocaleTimeString();

        messageDiv.appendChild(content);
        messageDiv.appendChild(time);

        messagesList.appendChild(messageDiv);
    });

    // Faire défiler vers le bas
    messagesList.scrollTop = messagesList.scrollHeight;
}

// Mise à jour de la liste des utilisateurs en ligne
function updateOnlineUsersList() {
    const onlineUsers = document.getElementById('online-users');
    const usersList = document.getElementById('users-list');

    if (!state.onlineUsers || state.onlineUsers.length === 0) {
        onlineUsers.innerHTML = '<li class="empty">Aucun utilisateur en ligne.</li>';
        usersList.innerHTML = '<div class="empty">Aucun utilisateur disponible.</div>';
        return;
    }

    // Trier les utilisateurs : d'abord ceux avec qui l'utilisateur a parlé récemment
    // puis par ordre alphabétique
    // La liste arrive déjà triée de l'API

    // Mettre à jour la liste des utilisateurs en ligne dans la sidebar
    onlineUsers.innerHTML = '';

    state.onlineUsers.forEach(user => {
        // Ne pas afficher l'utilisateur courant
        if (state.currentUser && user.id === state.currentUser.id) {
            return;
        }

        const li = document.createElement('li');
        li.onclick = () => navigateTo('messages', { user });

        const status = document.createElement('span');
        status.className = `online-status ${user.online ? 'online' : 'offline'}`;

        const username = document.createElement('span');
        username.textContent = user.username;

        li.appendChild(status);
        li.appendChild(username);

        onlineUsers.appendChild(li);
    });

    // Mettre à jour la liste des utilisateurs dans la page de messages
    usersList.innerHTML = '';

    state.onlineUsers.forEach(user => {
        // Ne pas afficher l'utilisateur courant
        if (state.currentUser && user.id === state.currentUser.id) {
            return;
        }

        const li = document.createElement('li');
        li.onclick = () => navigateTo('messages', { user });

        // Marquer l'utilisateur actuel comme actif
        if (state.currentChatUser && user.id === state.currentChatUser.id) {
            li.className = 'active';
        }

        const status = document.createElement('span');
        status.className = `online-status ${user.online ? 'online' : 'offline'}`;

        const username = document.createElement('span');
        username.textContent = user.username;

        li.appendChild(status);
        li.appendChild(username);

        usersList.appendChild(li);
    });
}

// Initialiser l'application
document.addEventListener('DOMContentLoaded', () => {
    console.log("DOM entièrement chargé, initialisation de l'application...");
    initApp();
});

// Gérer la navigation du navigateur (boutons précédent/suivant)
window.addEventListener('popstate', (event) => {
    if (event.state && event.state.page) {
        navigateTo(event.state.page, event.state.data);
    } else {
        navigateTo('home');
    }
});

// Exporter les fonctions et l'état pour les autres modules
export { state, updateAppState, navigateTo };
