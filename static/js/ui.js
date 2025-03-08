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
    // Récupérer tous les liens de navigation
    const navLinks = document.querySelectorAll('.nav-link');
    
    // Ajouter un événement de clic à chaque lien
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            
            // Récupérer la page cible
            const page = link.getAttribute('data-page');
            
            // Naviguer vers cette page
            navigateTo(page);
        });
    });
}

// Configurer les modals
function setupModals() {
    // Login Modal
    const loginButton = document.getElementById('login-button');
    const loginModal = document.getElementById('login-modal');
    const loginClose = loginModal.querySelector('.close');
    
    loginButton.addEventListener('click', () => {
        loginModal.style.display = 'block';
    });
    
    loginClose.addEventListener('click', () => {
        loginModal.style.display = 'none';
    });
    
    // Register Modal
    const registerButton = document.getElementById('register-button');
    const registerModal = document.getElementById('register-modal');
    const registerClose = registerModal.querySelector('.close');
    
    registerButton.addEventListener('click', () => {
        registerModal.style.display = 'block';
    });
    
    registerClose.addEventListener('click', () => {
        registerModal.style.display = 'none';
    });
    
    // Fermer les modals en cliquant en dehors
    window.addEventListener('click', (event) => {
        if (event.target === loginModal) {
            loginModal.style.display = 'none';
        }
        if (event.target === registerModal) {
            registerModal.style.display = 'none';
        }
    });
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
    // Masquer toutes les pages
    const pages = document.querySelectorAll('.page');
    pages.forEach(page => {
        page.classList.remove('active');
    });
    
    // Afficher la page demandée
    const page = document.getElementById(`${pageId}-page`);
    if (page) {
        page.classList.add('active');
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
