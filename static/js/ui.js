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

    if (loginButton && loginModal) {
        console.log("Bouton et modal de connexion trouvés");
        const loginClose = loginModal.querySelector('.close');

        loginButton.addEventListener('click', () => {
            console.log("Clic sur le bouton de connexion");
            loginModal.style.display = 'block';
        });

        if (loginClose) {
            loginClose.addEventListener('click', () => {
                loginModal.style.display = 'none';
            });
        }
    } else {
        console.error("Bouton ou modal de connexion non trouvé");
    }

    // Register Modal
    const registerButton = document.getElementById('register-button');
    const registerModal = document.getElementById('register-modal');

    if (registerButton && registerModal) {
        console.log("Bouton et modal d'inscription trouvés");
        const registerClose = registerModal.querySelector('.close');

        registerButton.addEventListener('click', () => {
            console.log("Clic sur le bouton d'inscription");
            registerModal.style.display = 'block';
        });

        if (registerClose) {
            registerClose.addEventListener('click', () => {
                registerModal.style.display = 'none';
            });
        }
    } else {
        console.error("Bouton ou modal d'inscription non trouvé");
    }

    // New Post Modal
    const newPostButton = document.getElementById('new-post-button');
    const newPostModal = document.getElementById('new-post-modal');

    if (newPostButton && newPostModal) {
        console.log("Bouton et modal de nouvelle publication trouvés");
        const newPostClose = newPostModal.querySelector('.close');

        newPostButton.addEventListener('click', () => {
            console.log("Clic sur le bouton de nouvelle publication");
            newPostModal.style.display = 'block';
        });

        if (newPostClose) {
            newPostClose.addEventListener('click', () => {
                newPostModal.style.display = 'none';
            });
        }
    }

    // Fermer les modals en cliquant en dehors
    window.addEventListener('click', (event) => {
        if (loginModal && event.target === loginModal) {
            loginModal.style.display = 'none';
        }
        if (registerModal && event.target === registerModal) {
            registerModal.style.display = 'none';
        }
        if (newPostModal && event.target === newPostModal) {
            newPostModal.style.display = 'none';
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

    if (!loginButton || !registerButton || !userInfo || !username) {
        console.error("Certains éléments d'authentification n'ont pas été trouvés");
        return;
    }

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