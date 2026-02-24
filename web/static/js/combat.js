const CombatUI = {
    show(combatData) {
        const hud = document.getElementById('combat-hud');
        hud.classList.remove('hidden');

        if (!combatData) return;

        // Turn counter
        document.getElementById('combat-turn').textContent = 'Turn ' + combatData.turn;

        // Player bars
        this.updateBar('player-hp', combatData.player_hp, combatData.player_max_hp);
        this.updateBar('player-mp', combatData.player_mp, combatData.player_max_mp);
        this.updateBar('player-sp', combatData.player_sp, combatData.player_max_sp);

        // Monster
        document.getElementById('monster-name').textContent = combatData.monster_name;
        this.updateBar('monster-hp', combatData.monster_hp, combatData.monster_max_hp);
        this.updateBar('monster-mp', combatData.monster_mp, combatData.monster_max_mp);
        this.updateBar('monster-sp', combatData.monster_sp, combatData.monster_max_sp);

        // Effects
        this.renderEffects('player-effects', combatData.player_effects || []);
        this.renderEffects('monster-effects', combatData.monster_effects || []);

        // Guards
        this.renderGuards(combatData.guards || []);
    },

    hide() {
        document.getElementById('combat-hud').classList.add('hidden');
    },

    updateBar(prefix, current, max) {
        const pct = max > 0 ? Math.max(0, Math.min(100, (current / max) * 100)) : 0;
        const barFill = document.getElementById(prefix + '-bar');
        const barText = document.getElementById(prefix + '-text');
        if (barFill) barFill.style.width = pct + '%';
        if (barText) barText.textContent = current + '/' + max;
    },

    renderEffects(elementId, effects) {
        const container = document.getElementById(elementId);
        container.innerHTML = '';
        effects.forEach(eff => {
            const badge = document.createElement('span');
            badge.className = 'effect-badge effect-' + eff.name.replace(/\s+/g, '_');
            badge.textContent = eff.name + ':' + eff.duration;
            container.appendChild(badge);
        });
    },

    renderGuards(guards) {
        const panel = document.getElementById('guard-panel');
        const list = document.getElementById('guard-list');
        if (!guards || guards.length === 0) {
            panel.classList.add('hidden');
            return;
        }
        panel.classList.remove('hidden');
        list.innerHTML = '';
        guards.forEach(g => {
            const div = document.createElement('div');
            div.className = 'guard-entry' + (g.injured ? ' injured' : '');
            div.textContent = g.name + ' HP:' + g.hp + '/' + g.max_hp + (g.injured ? ' [INJURED]' : '');
            list.appendChild(div);
        });
    }
};
