import { fetchCategories, updatePostsList } from './app.js';

// Initialiser le module des publications
export function initPosts(state, updateAppState) {
    try {
        // Configurer le formulaire de création de publication
        setupNewPostForm(state, updateAppState);

        // Configurer le formulaire de commentaire
        setupCommentForm(state);

        // Configurer la visualisation des détails de publication
        setupPostDetail(state);
    } catch (error) {
        console.error('Erreur lors de l\'initialisation des publications:', error);
    }
}

// Configurer le formulaire de création de publication
function setupNewPostForm(state, updateAppState) {
    const newPostForm = document.getElementById('new-post-form');
    if (!newPostForm) return;

    // Charger les catégories dans le formulaire si elles ne sont pas déjà chargées
    const postCategorySelect = document.getElementById('post-category');
    if (postCategorySelect && postCategorySelect.options.length === 0 && state.categories.length > 0) {
        state.categories.forEach(category => {
            const option = document.createElement('option');
            option.value = category.id;
            option.textContent = category.name;
            postCategorySelect.appendChild(option);
        });
    } else if (postCategorySelect && postCategorySelect.options.length === 0) {
        // Si les catégories ne sont pas encore chargées, les charger
        fetchCategories();
    }

    // Soumettre le formulaire
    newPostForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        // Vérifier que l'utilisateur est authentifié
        if (!state.isAuthenticated) {
            alert('Vous devez être connecté pour créer une publication.');
            return;
        }

        const postData = {
            title: document.getElementById('post-title').value,
            content: document.getElementById('post-content').value,
            categoryId: parseInt(document.getElementById('post-category').value)
        };

        try {
            const response = await fetch('/api/posts', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(postData)
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }

            const post = await response.json();

            // Ajouter la nouvelle publication à la liste
            const posts = [post, ...state.posts];
            updateAppState({ posts });

            // Fermer le modal
            const newPostModal = document.getElementById('new-post-modal');
            if (newPostModal) {
                newPostModal.style.display = 'none';
            }

            // Réinitialiser le formulaire
            newPostForm.reset();

            // Si l'utilisateur est sur la page d'accueil, mettre à jour la liste des publications
            if (state.currentPage === 'home') {
                updatePostsList();
            }

            // Envoyer un message WebSocket pour informer les autres utilisateurs
            if (state.socket && state.socket.readyState === WebSocket.OPEN) {
                const message = {
                    type: 'post_created',
                    payload: post
                };
                state.socket.send(JSON.stringify(message));
            }
        } catch (error) {
            console.error('Erreur lors de la création de la publication:', error);
            alert('Erreur lors de la création de la publication: ' + error.message);
        }
    });
}

// Configurer le formulaire de commentaire
function setupCommentForm(state) {
    const commentForm = document.getElementById('comment-form');
    if (!commentForm) return;

    const commentInput = document.getElementById('comment-input');
    const postCommentButton = document.getElementById('post-comment');

    if (!commentInput || !postCommentButton) return;

    postCommentButton.addEventListener('click', async () => {
        // Vérifier que l'utilisateur est authentifié
        if (!state.isAuthenticated) {
            alert('Vous devez être connecté pour commenter.');
            return;
        }

        // Vérifier qu'une publication est sélectionnée
        if (!state.currentPost) {
            alert('Aucune publication sélectionnée.');
            return;
        }

        // Vérifier que le commentaire n'est pas vide
        const content = commentInput.value.trim();
        if (!content) {
            alert('Le commentaire ne peut pas être vide.');
            return;
        }

        try {
            const response = await fetch(`/api/posts/${state.currentPost.id}/comments`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content })
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }

            const comment = await response.json();

            // Ajouter le commentaire à la liste
            const commentsList = document.getElementById('comments-list');
            if (!commentsList) return;

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

            const commentContent = document.createElement('div');
            commentContent.className = 'comment-content';
            commentContent.textContent = comment.content;

            header.appendChild(author);
            header.appendChild(date);
            commentDiv.appendChild(header);
            commentDiv.appendChild(commentContent);

            // Supprimer le message "Aucun commentaire" s'il existe
            const emptyMessage = commentsList.querySelector('.empty');
            if (emptyMessage) {
                commentsList.removeChild(emptyMessage);
            }

            commentsList.appendChild(commentDiv);

            // Réinitialiser le formulaire
            commentInput.value = '';

            // Envoyer un message WebSocket pour informer les autres utilisateurs
            if (state.socket && state.socket.readyState === WebSocket.OPEN) {
                const message = {
                    type: 'comment_created',
                    payload: comment
                };
                state.socket.send(JSON.stringify(message));
            }
        } catch (error) {
            console.error('Erreur lors de la création du commentaire:', error);
            alert('Erreur lors de la création du commentaire: ' + error.message);
        }
    });
}

// Configurer la visualisation des détails de publication
function setupPostDetail(state) {
    // Cette fonction est appelée quand on navigue vers la page de détail d'une publication
    const postDetailPage = document.getElementById('post-detail-page');
    if (!postDetailPage) return;

    // Observer les changements de classe sur la page de détail
    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            if (mutation.type === 'attributes' &&
                mutation.attributeName === 'class' &&
                mutation.target.id === 'post-detail-page') {

                // Si la page de détail devient active
                if (mutation.target.classList.contains('active') && state.currentPost) {
                    // Afficher les détails de la publication
                    const postDetail = document.getElementById('post-detail');
                    if (!postDetail) return;

                    postDetail.innerHTML = `
                        <h1>${state.currentPost.title}</h1>
                        <div class="post-meta">
                            <div class="post-category">${state.currentPost.category}</div>
                            <div class="post-author">Par ${state.currentPost.username}</div>
                            <div class="post-date">${new Date(state.currentPost.createdAt).toLocaleString()}</div>
                        </div>
                        <div class="post-content">${state.currentPost.content}</div>
                    `;
                }
            }
        });
    });

    observer.observe(postDetailPage, { attributes: true });
}