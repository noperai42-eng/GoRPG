const GameConnection = {
    ws: null,
    onMessage: null,
    onClose: null,
    onOpen: null,
    reconnectAttempts: 0,
    maxReconnects: 5,
    reconnectDelay: 2000,

    connect(token) {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${protocol}//${location.host}/ws/game?token=${encodeURIComponent(token)}`;

        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
            this.reconnectAttempts = 0;
            if (this.onOpen) this.onOpen();
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (this.onMessage) this.onMessage(data);
            } catch(e) {
                console.error('Failed to parse message:', e);
            }
        };

        this.ws.onclose = (event) => {
            if (this.onClose) this.onClose(event);
            // Auto-reconnect unless intentionally closed
            if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnects) {
                this.reconnectAttempts++;
                setTimeout(() => this.connect(token), this.reconnectDelay * this.reconnectAttempts);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },

    sendCommand(type, value) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ type, value: value || '' }));
        }
    },

    disconnect() {
        if (this.ws) {
            this.ws.close(1000, 'user disconnect');
            this.ws = null;
        }
    },

    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN;
    }
};
