const GameUI = {
    messageLog: null,
    optionsPanel: null,
    promptPanel: null,

    init() {
        this.messageLog = document.getElementById('message-log');
        this.optionsPanel = document.getElementById('options-panel');
        this.promptPanel = document.getElementById('prompt-panel');

        document.getElementById('prompt-form').addEventListener('submit', (e) => {
            e.preventDefault();
            const input = document.getElementById('prompt-input');
            const value = input.value.trim();
            if (value) {
                GameConnection.sendCommand('input', value);
                input.value = '';
            }
        });
    },

    renderResponse(resp) {
        // Render messages
        if (resp.messages && resp.messages.length > 0) {
            resp.messages.forEach(msg => {
                const div = document.createElement('div');
                div.className = 'log-entry log-' + (msg.category || 'narrative');
                div.textContent = msg.text;
                this.messageLog.appendChild(div);
            });
            // Auto-scroll to bottom
            this.messageLog.scrollTop = this.messageLog.scrollHeight;
        }

        // Update player info in header
        if (resp.state && resp.state.player) {
            const p = resp.state.player;
            document.getElementById('player-info').textContent =
                p.name + ' | Lv.' + p.level + ' | HP:' + p.hp + '/' + p.max_hp;
        }

        // Show/hide combat HUD
        if (resp.state && resp.state.combat) {
            CombatUI.show(resp.state.combat);
        } else {
            CombatUI.hide();
        }

        // Render options (as buttons)
        this.optionsPanel.innerHTML = '';
        if (resp.options && resp.options.length > 0 && !resp.prompt) {
            resp.options.forEach(opt => {
                const btn = document.createElement('button');
                btn.className = 'option-btn' + (opt.enabled === false ? ' disabled' : '');
                btn.textContent = opt.label;
                btn.disabled = (opt.enabled === false);
                btn.addEventListener('click', () => {
                    if (opt.enabled !== false) {
                        GameConnection.sendCommand('select', opt.key);
                    }
                });
                this.optionsPanel.appendChild(btn);
            });
        }

        // Show/hide prompt
        if (resp.prompt) {
            this.promptPanel.classList.remove('hidden');
            document.getElementById('prompt-text').textContent = resp.prompt;
            document.getElementById('prompt-input').focus();
            this.optionsPanel.innerHTML = '';
        } else {
            this.promptPanel.classList.add('hidden');
        }

        // Handle exit
        if (resp.type === 'exit') {
            this.addSystemMessage('Game saved. Disconnecting...');
            setTimeout(() => {
                GameConnection.disconnect();
                App.showScreen('characters');
            }, 1500);
        }
    },

    addSystemMessage(text) {
        const div = document.createElement('div');
        div.className = 'log-entry log-system';
        div.textContent = text;
        this.messageLog.appendChild(div);
        this.messageLog.scrollTop = this.messageLog.scrollHeight;
    },

    clearLog() {
        this.messageLog.innerHTML = '';
    }
};
