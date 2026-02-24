const Auth = {
    token: null,
    username: null,

    init() {
        // Check localStorage for existing token
        this.token = localStorage.getItem('token');
        this.username = localStorage.getItem('username');
    },

    isLoggedIn() {
        return !!this.token;
    },

    async register(username, password) {
        const resp = await fetch('/api/register', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username, password})
        });
        const data = await resp.json();
        if (!resp.ok) throw new Error(data.error || 'Registration failed');
        return data;
    },

    async login(username, password) {
        const resp = await fetch('/api/login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username, password})
        });
        const data = await resp.json();
        if (!resp.ok) throw new Error(data.error || 'Login failed');
        this.token = data.token;
        this.username = data.username;
        localStorage.setItem('token', this.token);
        localStorage.setItem('username', this.username);
        return data;
    },

    logout() {
        this.token = null;
        this.username = null;
        localStorage.removeItem('token');
        localStorage.removeItem('username');
    },

    getAuthHeaders() {
        return { 'Authorization': 'Bearer ' + this.token, 'Content-Type': 'application/json' };
    }
};
