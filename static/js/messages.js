// fichier: static/js/messages.js

// Variables pour l'indicateur de frappe
let typingTimer;
let isTyping = false;
const TYPING_DELAY = 1000; // Délai avant de considérer que l'utilisateur a arrêté de taper

// Initialiser le module des messages
export function initMessages(state) {
    // Configurer le formulaire d'envoi de message
    setupMessageForm(state);
    
    // Configurer l'indicateur de frappe
    setupTypingIndicator(state);
    
    // Configurer la visualisation des messages
    setupMessagesView(state);
}

// Configurer le formulaire d'envoi de message
function setupMessageForm(state) {
    const messageForm = document.getElementById('message-form');
    const messageInput = document.getElementById('message-input');
    const sendButton = document.getElementById('send-message');
    
    sendButton.addEventListener('click', () => {
        sendMessage(state);
    });
    
    messageInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage(state);
        }
    });
}

// Envoyer un message
async function sendMessage(state) {
    // Vérifier que l'utilisateur est authentifié
    if (!state.isAuthenticated) {
        alert('Vous devez être connecté pour envoyer des messages.');
        return;
    }
    
    // Vérifier qu'un destinataire est sélectionné
    if (!state.currentChatUser) {
        alert('Veuillez sélectionner un utilisateur pour discuter.');
        return;
    }
    
    const messageInput = document.getElementById('message-input');
    const content = messageInput.value.trim();
    
    // Vérifier que le message n'est pas vide
    if (!content) {
        return;
    }
    
    try {
        // Préparer les données du message
        const messageData = {
            receiverId: state.currentChatUser.id,
            content: content
        };
        
        // Désactiver temporairement le formulaire
        messageInput.disabled = true;
        
        // Réinitialiser l'indicateur de frappe
        resetTypingIndicator(state);
        
        // Si l'utilisateur utilise le WebSocket, envoyer le message via WebSocket
        if (state.socket && state.socket.readyState === WebSocket.OPEN) {
            const wsMessage = {
                type: 'private_message',
                payload: messageData
            };
            state.socket.send(JSON.stringify(wsMessage));
            
            // Réinitialiser le formulaire
            messageInput.value = '';
            messageInput.disabled = false;
            messageInput.focus();
        } else {
            // Sinon, envoyer le message via API REST
            const response = await fetch('/api/messages', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(messageData)
            });
            
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
            
            const message = await response.json();
            
            // Ajouter le message à la liste
            const messagesList = document.getElementById('messages-list');
            
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message sent';
            
            const content = document.createElement('div');
            content.className = 'message-content';
            content.textContent = message.content;
            
            const time = document.createElement('div');
            time.className = 'message-time';
            time.textContent = new Date(message.createdAt).toLocaleTimeString();
            
            messageDiv.appendChild(content);
            messageDiv.appendChild(time);
            
            // Supprimer le message "Aucun message" s'il existe
            const emptyMessage = messagesList.querySelector('.empty');
            if (emptyMessage) {
                messagesList.removeChild(emptyMessage);
            }
            
            messagesList.appendChild(messageDiv);
            
            // Faire défiler vers le bas
            messagesList.scrollTop = messagesList.scrollHeight;
            
            // Réinitialiser le formulaire
            messageInput.value = '';
            messageInput.disabled = false;
            messageInput.focus();
        }
    } catch (error) {
        console.error('Erreur lors de l\'envoi du message:', error);
        alert('Erreur lors de l\'envoi du message: ' + error.message);
        
        // Réactiver le formulaire
        messageInput.disabled = false;
    }
}

// Configurer l'indicateur de frappe
function setupTypingIndicator(state) {
    const messageInput = document.getElementById('message-input');
    
    messageInput.addEventListener('input', () => {
        // Si l'utilisateur n'est pas en train de taper
        if (!isTyping) {
            isTyping = true;
            
            // Envoyer l'indicateur de frappe
            sendTypingIndicator(state, true);
        }
        
        // Réinitialiser le timer
        clearTimeout(typingTimer);
        
        // Définir un nouveau timer
        typingTimer = setTimeout(() => {
            isTyping = false;
            
            // Envoyer l'indicateur de frappe
            sendTypingIndicator(state, false);
        }, TYPING_DELAY);
    });
}

// Envoyer l'indicateur de frappe
function sendTypingIndicator(state, isTyping) {
    // Vérifier que l'utilisateur est authentifié
    if (!state.isAuthenticated) {
        return;
    }
    
    // Vérifier qu'un destinataire est sélectionné
    if (!state.currentChatUser) {
        return;
    }
    
    // Préparer les données de l'indicateur
    const typingData = {
        targetUserId: state.currentChatUser.id,
        isTyping: isTyping
    };
    
    // Si l'utilisateur utilise le WebSocket, envoyer l'indicateur via WebSocket
    if (state.socket && state.socket.readyState === WebSocket.OPEN) {
        const wsMessage = {
            type: 'typing_indicator',
            payload: typingData
        };
        state.socket.send(JSON.stringify(wsMessage));
    } else {
        // Sinon, envoyer l'indicateur via API REST
        fetch('/api/typing', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(typingData)
        }).catch(error => {
            console.error('Erreur lors de l\'envoi de l\'indicateur de frappe:', error);
        });
    }
}

// Réinitialiser l'indicateur de frappe
function resetTypingIndicator(state) {
    isTyping = false;
    clearTimeout(typingTimer);
    
    // Envoyer l'indicateur de frappe
    sendTypingIndicator(state, false);
}

// Configurer la visualisation des messages
function setupMessagesView(state) {
    // Cette fonction est appelée quand on navigue vers la page des messages
    const observer = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
            if (mutation.type === 'attributes' && 
                mutation.attributeName === 'class' &&
                mutation.target.id === 'messages-page') {
                
                // Si la page des messages devient active
                if (mutation.target.classList.contains('active')) {
                    // Mettre à jour l'en-tête de la conversation
                    updateChatHeader(state);
                    
                    // Afficher ou masquer le formulaire de message
                    if (state.currentChatUser) {
                        document.getElementById('message-form').classList.remove('hidden');
                    } else {
                        document.getElementById('message-form').classList.add('hidden');
                    }
                    
                    // Masquer l'indicateur de frappe
                    document.getElementById('typing-indicator').classList.add('hidden');
                }
            }
        });
    });
    
    // Observer les changements de classe sur la page des messages
    const messagesPage = document.getElementById('messages-page');
    observer.observe(messagesPage, { attributes: true });
}

// Mettre à jour l'en-tête de la conversation
function updateChatHeader(state) {
    const chatHeader = document.getElementById('chat-header');
    
    if (state.currentChatUser) {
        chatHeader.innerHTML = `
            <h2>Conversation avec ${state.currentChatUser.username}</h2>
        `;
    } else {
        chatHeader.innerHTML = `
            <h2>Sélectionnez un utilisateur pour discuter</h2>
        `;
    }
}
