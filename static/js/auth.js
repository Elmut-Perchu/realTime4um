// Initialiser le module d'authentification
export async function initAuth(state, updateAppState) {
    try {
        // Vérifier si l'utilisateur est déjà connecté
        const user = await checkAuth();
        if (user) {
            updateAppState({
                currentUser: user,
                isAuthenticated: true
            });
        }

        // Configurer les événements pour les formulaires d'authentification
        setupAuthForms(updateAppState);
    } catch (error) {
        console.error('Erreur lors de l\'initialisation de l\'authentification:', error);
    }
}

// Vérifier si l'utilisateur est authentifié
export function isAuthenticated() {
    return document.cookie.includes('session_id=');
}

// Récupérer l'utilisateur actuel
export async function getCurrentUser() {
    try {
        const response = await fetch('/api/me');
        if (!response.ok) {
            return null;
        }

        return await response.json();
    } catch (error) {
        console.error('Erreur lors de la récupération de l\'utilisateur:', error);
        return null;
    }
}

// Vérifier l'authentification
async function checkAuth() {
    if (isAuthenticated()) {
        return await getCurrentUser();
    }
    return null;
}

// Analyser l'erreur de réponse
async function parseResponseError(response) {
    try {
        // Essayer d'analyser comme JSON d'abord
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            const errorJson = await response.json();
            return errorJson.message || JSON.stringify(errorJson);
        } else {
            // Sinon, obtenir le texte brut
            const text = await response.text();
            return text;
        }
    } catch (e) {
        console.error('Erreur lors de l\'analyse de l\'erreur de réponse:', e);
        return "Une erreur inconnue s'est produite";
    }
}

// Configurer les formulaires d'authentification
function setupAuthForms(updateAppState) {
    // Formulaire de connexion
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const identifier = document.getElementById('login-identifier').value;
            const password = document.getElementById('login-password').value;

            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ identifier, password })
                });

                if (!response.ok) {
                    const errorMsg = await parseResponseError(response);
                    throw new Error(errorMsg);
                }

                const data = await response.json();

                // Mettre à jour l'état
                updateAppState({
                    currentUser: data.user,
                    isAuthenticated: true
                });

                // Fermer le modal
                const loginModal = document.getElementById('login-modal');
                if (loginModal) {
                    loginModal.style.display = 'none';
                }

                // Réinitialiser le formulaire
                loginForm.reset();

                // Rediriger vers la page d'accueil si ce n'est pas déjà le cas
                if (window.location.pathname !== '/') {
                    window.location.href = '/';
                }
            } catch (error) {
                console.error('Erreur de connexion:', error);
                const errorMessage = error.message || "Erreur de connexion";
                alert(errorMessage);
            }
        });
    }

    // Formulaire d'inscription
    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const userData = {
                username: document.getElementById('register-username').value,
                age: parseInt(document.getElementById('register-age').value),
                gender: document.getElementById('register-gender').value,
                firstName: document.getElementById('register-firstname').value,
                lastName: document.getElementById('register-lastname').value,
                email: document.getElementById('register-email').value,
                password: document.getElementById('register-password').value
            };

            try {
                const response = await fetch('/api/register', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(userData)
                });

                if (!response.ok) {
                    const errorMsg = await parseResponseError(response);
                    throw new Error(errorMsg);
                }

                const data = await response.json();

                // Mettre à jour l'état
                updateAppState({
                    currentUser: data,
                    isAuthenticated: true
                });

                // Fermer le modal
                const registerModal = document.getElementById('register-modal');
                if (registerModal) {
                    registerModal.style.display = 'none';
                }

                // Réinitialiser le formulaire
                registerForm.reset();

                // Rediriger vers la page d'accueil si ce n'est pas déjà le cas
                if (window.location.pathname !== '/') {
                    window.location.href = '/';
                }
            } catch (error) {
                console.error('Erreur d\'inscription:', error);
                const errorMessage = error.message || "Erreur d'inscription";
                alert(errorMessage);
            }
        });
    }

    // Bouton de déconnexion
    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
        logoutButton.addEventListener('click', async () => {
            try {
                const response = await fetch('/api/logout', { method: 'POST' });

                if (!response.ok) {
                    const errorMsg = await parseResponseError(response);
                    throw new Error(errorMsg);
                }

                // Mettre à jour l'état
                updateAppState({
                    currentUser: null,
                    isAuthenticated: false,
                    currentPage: 'home'
                });

                // Rediriger vers la page d'accueil
                window.location.href = '/';
            } catch (error) {
                console.error('Erreur de déconnexion:', error);
                const errorMessage = error.message || "Erreur de déconnexion";
                alert(errorMessage);
            }
        });
    }
}