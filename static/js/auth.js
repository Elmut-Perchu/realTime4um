// Initialiser le module d'authentification
export async function initAuth(state, updateAppState) {
    try {
        console.log("Vérification de l'authentification...");

        // Vérifier si l'utilisateur est déjà connecté
        if (isAuthenticated()) {
            console.log("Token de session trouvé dans localStorage, tentative de récupération de l'utilisateur");
            const user = await getCurrentUser();
            if (user) {
                console.log("Utilisateur authentifié:", user.username);
                updateAppState({
                    currentUser: user,
                    isAuthenticated: true
                });
            } else {
                console.log("Token de session présent mais utilisateur non trouvé");
                // Supprimer le token obsolète
                localStorage.removeItem('session_id');
            }
        } else {
            console.log("Aucun token de session trouvé dans localStorage");
        }

        // Configurer les événements pour les formulaires d'authentification
        setupAuthForms(updateAppState);
    } catch (error) {
        console.error('Erreur lors de l\'initialisation de l\'authentification:', error);
    }
}

// Vérifier si l'utilisateur est authentifié
export function isAuthenticated() {
    return localStorage.getItem('session_id') !== null;
}

// Récupérer l'utilisateur actuel
export async function getCurrentUser() {
    try {
        console.log("Tentative de récupération de l'utilisateur actuel...");
        const sessionId = localStorage.getItem('session_id');

        if (!sessionId) {
            console.log("Aucun token de session trouvé");
            return null;
        }

        const response = await fetch('/api/me', {
            headers: {
                'Authorization': `Bearer ${sessionId}`
            }
        });

        if (!response.ok) {
            if (response.status === 401) {
                console.log("L'utilisateur n'est pas authentifié");
                // Nettoyer le localStorage si le token n'est plus valide
                localStorage.removeItem('session_id');
            } else {
                console.error("Erreur lors de la récupération de l'utilisateur:", response.status);
            }
            return null;
        }

        const user = await response.json();
        console.log("Utilisateur récupéré:", user.username);
        return user;
    } catch (error) {
        console.error('Erreur lors de la récupération de l\'utilisateur:', error);
        return null;
    }
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
                console.log("Tentative de connexion pour:", identifier);

                // Désactiver le formulaire pendant la requête
                Array.from(loginForm.elements).forEach(el => el.disabled = true);

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
                console.log("Connexion réussie pour:", data.user.username);

                // Stocker le token dans localStorage
                localStorage.setItem('session_id', data.sessionId);

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

                // Recharger la page complètement pour s'assurer que tout est mis à jour
                window.location.reload();
            } catch (error) {
                console.error('Erreur de connexion:', error);
                const errorMessage = error.message || "Erreur de connexion";
                alert(errorMessage);

                // Réactiver le formulaire
                Array.from(loginForm.elements).forEach(el => el.disabled = false);
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
                console.log("Tentative d'inscription pour:", userData.username);

                // Désactiver le formulaire pendant la requête
                Array.from(registerForm.elements).forEach(el => el.disabled = true);

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
                console.log("Inscription réussie pour:", data.username);

                // Récupérer l'ID de session de la réponse ou des en-têtes
                const sessionId = response.headers.get('X-Session-ID') || data.sessionId;
                if (sessionId) {
                    localStorage.setItem('session_id', sessionId);
                } else {
                    console.warn("Pas d'ID de session trouvé dans la réponse");
                }

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

                // Recharger la page complètement pour s'assurer que tout est mis à jour
                window.location.reload();
            } catch (error) {
                console.error('Erreur d\'inscription:', error);
                const errorMessage = error.message || "Erreur d'inscription";
                alert(errorMessage);

                // Réactiver le formulaire
                Array.from(registerForm.elements).forEach(el => el.disabled = false);
            }
        });
    }

    // Bouton de déconnexion
    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
        logoutButton.addEventListener('click', async () => {
            try {
                console.log("Tentative de déconnexion");
                const sessionId = localStorage.getItem('session_id');

                const response = await fetch('/api/logout', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${sessionId}`
                    }
                });

                if (!response.ok) {
                    const errorMsg = await parseResponseError(response);
                    throw new Error(errorMsg);
                }

                console.log("Déconnexion réussie");

                // Supprimer le token du localStorage
                localStorage.removeItem('session_id');

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