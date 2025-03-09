// Initialiser la connexion WebSocket
export function initWebSocket(state, messageHandler) {
    try {
        // Récupérer le token de session
        const sessionId = getSessionIdFromCookie();

        if (!sessionId) {
            console.error('Aucun token de session trouvé pour WebSocket');
            return null;
        }

        // Créer l'URL WebSocket
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?token=${sessionId}`;

        console.log('Tentative de connexion WebSocket à', wsUrl);

        // Créer la connexion WebSocket
        const socket = new WebSocket(wsUrl);

        // Configurer les événements
        socket.onopen = () => {
            console.log('Connexion WebSocket établie');
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