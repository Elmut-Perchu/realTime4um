// Initialiser la connexion WebSocket
export function initWebSocket(state, messageHandler) {
    try {
        // Ne pas initialiser le WebSocket si l'utilisateur n'est pas authentifié
        if (!state.isAuthenticated || !state.currentUser) {
            console.log('Utilisateur non authentifié, WebSocket non initialisé');
            return null;
        }

        // Récupérer le token de session depuis localStorage
        const sessionId = localStorage.getItem('session_id');

        if (!sessionId) {
            console.error('Aucun token de session trouvé dans localStorage pour WebSocket');
            return null;
        }

        // Créer l'URL WebSocket avec le token
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?ls_token=${sessionId}`;

        console.log('Tentative de connexion WebSocket à', wsUrl);

        // Créer la connexion WebSocket
        const socket = new WebSocket(wsUrl);

        // Configurer les événements
        socket.onopen = () => {
            console.log('Connexion WebSocket établie');

            // Envoyer un message test pour vérifier la connexion
            try {
                socket.send(JSON.stringify({
                    type: 'connection_test',
                    payload: { userId: state.currentUser.id }
                }));
                console.log('Message de test WebSocket envoyé');
            } catch (error) {
                console.error('Erreur lors de l\'envoi du message de test:', error);
            }
        };

        socket.onmessage = (event) => {
            try {
                console.log('Message WebSocket reçu:', event.data);
                messageHandler(event);
            } catch (error) {
                console.error('Erreur lors du traitement du message WebSocket:', error);
            }
        };

        socket.onclose = (event) => {
            console.log(`Connexion WebSocket fermée: ${event.code} ${event.reason}`);

            // Essayer de se reconnecter après un délai si la connexion n'est pas fermée volontairement
            if (event.code !== 1000 && event.code !== 1001) {
                setTimeout(() => {
                    if (state.isAuthenticated) {
                        console.log('Tentative de reconnexion WebSocket...');
                        state.socket = initWebSocket(state, messageHandler);
                    }
                }, 5000); // Reconnecter après 5 secondes
            }
        };

        socket.onerror = (error) => {
            console.error('Erreur WebSocket:', error);
        };

        return socket;
    } catch (error) {
        console.error('Erreur lors de l\'initialisation du WebSocket:', error);
        return null;
    }
}

// Récupérer l'ID de session à partir du cookie
function getSessionIdFromCookie() {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        cookie = cookie.trim();
        if (cookie.startsWith('session_id=')) {
            return cookie.substring('session_id='.length);
        }
    }
    return '';
}