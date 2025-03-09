// fichier: static/js/ui.js

// Initialiser l'interface utilisateur
export function initUI(state, navigateTo) {
    // Configurer la navigation
    setupNavigation(navigateTo);

    // Configurer les modals
    setupModals();

    // Mettre à jour l'interface en fonction de l'état initial
    updateUI(state);
}

// Configurer la navigation
function setupNavigation(navigateTo) {
    console.log("Configuration de la navigation...");

    // Récupérer tous les liens de navigation
    const navLinks = document.querySelectorAll('.nav-link');

    console.log("Liens de navigation trouvés:", navLinks.length);

    // Ajouter un événement de clic à chaque lien
    navLinks.forEach(link => {
        const page = link.getAttribute('data-page');
        console.log(`Configuration du lien vers ${page}`);

        link.addEventListener('click', (e) => {
            console.log(`Clic sur le lien vers ${page}`);
            e.preventDefault(); // Empêcher le comportement par défaut du lien

            // Naviguer vers cette page
            navigateTo(page);
        });
    });
}

// Configurer les modals
function setupModals() {
    console.log("Configuration des modals...");

    // Login Modal
    const loginButton = document.getElementById('login-button');
    const loginModal = document.getElementById('login-modal');

    if (!loginButton) {
        console.error("Bouton de connexion non trouvé (id: login-button)");
    } else {
        console.log("Bouton de connexion trouvé");
    }

    if (!loginModal) {
        console.error("Modal de connexion non trouvé (id: login-modal)");
    } else {
        console.log("Modal de connexion trouvé");
        const loginClose = loginModal.querySelector('.close');

        if (!loginClose) {
            console.error("Bouton de fermeture du modal de connexion non trouvé");
        }

        loginButton.addEventListener('click', () => {
            console.log("Clic sur le bouton de connexion");
            loginModal.style.display = 'block';
        });

        if (loginClose) {
            loginClose.addEventListener('click', () => {
                loginModal.style.display = 'none';
            });
        }
    }

    // Register Modal
    const registerButton = document.getElementById('register-button');
    const registerModal = document.getElementById('register-modal');

    if (!registerButton) {
        console.error("Bouton d'inscription non trouvé (id: register-button)");
    } else {
        console.log("Bouton d'inscription trouvé");
    }

    if (!registerModal) {
        console.error("Modal d'inscription non trouvé (id: register-modal)");
    } else {
        console.log("Modal d'inscription trouvé");
        const registerClose = registerModal.querySelector('.close');

        if (!registerClose) {
            console.error("Bouton de fermeture du modal d'inscription non trouvé");
        }

        registerButton.addEventListener('click', () => {
            console.log("Clic sur le bouton d'inscription");
            registerModal.style.display = 'block';
        });

        if (registerClose) {
            registerClose.addEventListener('click', () => {
                registerModal.style.display = 'none';
            });
        }
    }

    // Fermer les modals en cliquant en dehors
    window.addEventListener('click', (event) => {
        if (event.target === loginModal) {
            loginModal.style.display = 'none';
        }
        if (event.target === registerModal) {
            registerModal.style.display = 'none';
        }
    });

    console.log("Configuration des modals terminée");
}

// Mettre à jour l'interface utilisateur
export function updateUI(state) {
    // Mettre à jour l'affichage en fonction de l'authentification
    updateAuthUI(state);

    // Mettre à jour l'affichage des éléments sensibles à l'authentification
    updateAuthRequiredUI(state.isAuthenticated);
}

// Mettre à jour l'interface en fonction de l'authentification
function updateAuthUI(state) {
    const loginButton = document.getElementById('login-button');
    const registerButton = document.getElementById('register-button');
    const userInfo = document.getElementById('user-info');
    const username = document.getElementById('username');

    if (state.isAuthenticated && state.currentUser) {
        // Masquer les boutons de connexion et d'inscription
        loginButton.style.display = 'none';
        registerButton.style.display = 'none';

        // Afficher les informations de l'utilisateur
        userInfo.style.display = 'flex';
        username.textContent = state.currentUser.username;
    } else {
        // Afficher les boutons de connexion et d'inscription
        loginButton.style.display = 'block';
        registerButton.style.display = 'block';

        // Masquer les informations de l'utilisateur
        userInfo.style.display = 'none';
    }
}

// Mettre à jour l'affichage des éléments sensibles à l'authentification
function updateAuthRequiredUI(isAuthenticated) {
    const authRequiredElements = document.querySelectorAll('.auth-required');
    const authNotRequiredElements = document.querySelectorAll('.auth-not-required');

    authRequiredElements.forEach(element => {
        if (isAuthenticated) {
            element.classList.remove('hidden');
        } else {
            element.classList.add('hidden');
        }
    });

    authNotRequiredElements.forEach(element => {
        if (isAuthenticated) {
            element.classList.add('hidden');
        } else {
            element.classList.remove('hidden');
        }
    });
}

// Afficher une page
export function showPage(pageId) {
    console.log(`Affichage de la page: ${pageId}`);

    // Masquer toutes les pages
    const pages = document.querySelectorAll('.page');
    pages.forEach(page => {
        page.classList.remove('active');
    });

    // Afficher la page demandée
    const page = document.getElementById(`${pageId}-page`);
    if (page) {
        console.log(`Page ${pageId} trouvée et activée`);
        page.classList.add('active');
    } else {
        console.error(`Page ${pageId} non trouvée!`);
    }

    // Mettre à jour les liens de navigation
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        const linkPage = link.getAttribute('data-page');
        if (linkPage === pageId) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
}