const GameConnection = {
    ws: null,
    onOpen: null,
    onClose: null,
    reconnectAttempts: 0,
    maxReconnects: 5,
    reconnectDelay: 2000,
    _token: null,
    _messageQueue: [],
    _flushScheduled: false,

    connect(token) {
        this._token = token;
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
                // Queue messages and flush on next microtask to batch
                // rapid WebSocket messages into a single Alpine reactive cycle
                this._messageQueue.push(data);
                if (!this._flushScheduled) {
                    this._flushScheduled = true;
                    queueMicrotask(() => this._flushMessages());
                }
            } catch(e) {
                console.error('Failed to parse message:', e);
            }
        };

        this.ws.onclose = (event) => {
            if (this.onClose) this.onClose(event);
            // Auto-reconnect unless intentionally closed
            if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnects) {
                this.reconnectAttempts++;
                setTimeout(() => this.connect(this._token), this.reconnectDelay * this.reconnectAttempts);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },

    _flushMessages() {
        this._flushScheduled = false;
        const queue = this._messageQueue;
        this._messageQueue = [];
        if (!Alpine || !Alpine.store('game')) return;
        const store = Alpine.store('game');
        for (const data of queue) {
            try {
                store.handleResponse(data);
            } catch(e) {
                // Suppress Alpine x-for DOM race condition errors
                if (e instanceof TypeError) continue;
                console.error('Error handling message:', e);
            }
        }
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
