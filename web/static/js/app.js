const App = {
    currentScreen: 'auth',
    selectedCharacter: null,

    init() {
        Auth.init();
        GameUI.init();

        // Set up auth tab switching
        document.querySelectorAll('#auth-tabs .tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('#auth-tabs .tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                const tabName = tab.dataset.tab;
                document.getElementById('login-form').classList.toggle('hidden', tabName !== 'login');
                document.getElementById('register-form').classList.toggle('hidden', tabName !== 'register');
            });
        });

        // Login form
        document.getElementById('login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('login-username').value;
            const password = document.getElementById('login-password').value;
            const errorEl = document.getElementById('login-error');
            try {
                await Auth.login(username, password);
                errorEl.textContent = '';
                this.showScreen('characters');
                this.loadCharacters();
            } catch(err) {
                errorEl.textContent = err.message;
            }
        });

        // Register form
        document.getElementById('register-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('register-username').value;
            const password = document.getElementById('register-password').value;
            const errorEl = document.getElementById('register-error');
            try {
                await Auth.register(username, password);
                // Auto-login after registration
                await Auth.login(username, password);
                errorEl.textContent = '';
                this.showScreen('characters');
                this.loadCharacters();
            } catch(err) {
                errorEl.textContent = err.message;
            }
        });

        // Create character button
        document.getElementById('btn-create-char').addEventListener('click', () => this.createCharacter());

        // Play button
        document.getElementById('btn-play').addEventListener('click', () => this.startGame());

        // Logout button
        document.getElementById('btn-logout').addEventListener('click', () => {
            Auth.logout();
            this.showScreen('auth');
        });

        // Disconnect button
        document.getElementById('btn-disconnect').addEventListener('click', () => {
            GameConnection.disconnect();
            this.showScreen('characters');
        });

        // WebSocket callbacks
        GameConnection.onMessage = (data) => {
            GameUI.renderResponse(data);
        };

        GameConnection.onClose = () => {
            if (this.currentScreen === 'game') {
                GameUI.addSystemMessage('Connection lost.');
            }
        };

        // Auto-login if token exists
        if (Auth.isLoggedIn()) {
            this.showScreen('characters');
            this.loadCharacters();
        }
    },

    showScreen(name) {
        this.currentScreen = name;
        document.querySelectorAll('.screen').forEach(s => s.classList.remove('active'));
        document.getElementById('screen-' + name).classList.add('active');
    },

    async loadCharacters() {
        try {
            const resp = await fetch('/api/characters', {
                headers: Auth.getAuthHeaders()
            });
            if (resp.status === 401) {
                Auth.logout();
                this.showScreen('auth');
                return;
            }
            const data = await resp.json();
            this.renderCharacterList(data.characters || []);
        } catch(err) {
            document.getElementById('char-error').textContent = 'Failed to load characters';
        }
    },

    renderCharacterList(characters) {
        const list = document.getElementById('character-list');
        list.innerHTML = '';
        this.selectedCharacter = null;

        characters.forEach(name => {
            const div = document.createElement('div');
            div.className = 'char-item';
            div.textContent = name;
            div.addEventListener('click', () => {
                document.querySelectorAll('.char-item').forEach(c => c.classList.remove('selected'));
                div.classList.add('selected');
                this.selectedCharacter = name;
            });
            list.appendChild(div);
        });

        // Auto-select first character
        if (characters.length > 0) {
            list.firstChild.click();
        }
    },

    async createCharacter() {
        const nameInput = document.getElementById('new-char-name');
        const name = nameInput.value.trim();
        const errorEl = document.getElementById('char-error');

        if (!name) {
            errorEl.textContent = 'Enter a character name';
            return;
        }

        try {
            const resp = await fetch('/api/characters', {
                method: 'POST',
                headers: Auth.getAuthHeaders(),
                body: JSON.stringify({ name })
            });
            const data = await resp.json();
            if (!resp.ok) throw new Error(data.error || 'Failed to create character');
            errorEl.textContent = '';
            nameInput.value = '';
            this.loadCharacters();
        } catch(err) {
            errorEl.textContent = err.message;
        }
    },

    startGame() {
        if (!Auth.token) {
            this.showScreen('auth');
            return;
        }

        GameUI.clearLog();
        CombatUI.hide();
        this.showScreen('game');

        GameConnection.connect(Auth.token);
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => App.init());
