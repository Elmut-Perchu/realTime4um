// fichier: static/js/websocket.js

// Initialiser la connexion WebSocket
export function initWebSocket(state, messageHandler) {
    // Récupérer le token de session
    const sessionId = getSessionIdFromCookie();
    
    // Créer l'URL WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?token=${sessionId}`;
    
    // Créer la connexion WebSocket
    const socket = new WebSocket(wsUrl);
    
    // Configurer les événements
    socket.onopen = () => {
        console.log('Connexion WebSocket établie');
    };
    
    socket.onmessage = messageHandler;
    
    socket.onclose = (event) => {
        console.log(`Connexion WebSocket fermée: ${event.code} ${event.reason}`);
        
        // Essayer de se reconnecter après un délai si la connexion n'est pas fermée volontairement
        if (event.code !== 1000) {
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
