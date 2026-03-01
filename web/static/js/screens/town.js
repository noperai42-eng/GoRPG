// Town screen Alpine component
function townScreen() {
    return {
        subTab: 'inn',
        subTabs: [
            { id: 'inn', label: 'Inn' },
            { id: 'mayor', label: 'Mayor' },
            { id: 'quests', label: 'Fetch Quests' },
            { id: 'log', label: 'Attack Log' },
        ],

        get g() { return this.$store.game; },
        get t() { return this.$store.game.town; },
        get p() { return this.$store.game.player || {}; },
        get hasTown() { return this.t !== null; },

        get isTownScreen() {
            const s = this.g.serverScreen;
            return s && s.startsWith('town_');
        },

        get isMayor() {
            return this.t && this.t.is_current_mayor;
        },

        get guests() {
            return (this.t && this.t.guests) || [];
        },

        get mayor() {
            return this.t && this.t.mayor;
        },

        get fetchQuests() {
            return (this.t && this.t.fetch_quests) || [];
        },

        get attackLog() {
            return (this.t && this.t.attack_log) || [];
        },

        get treasury() {
            return (this.t && this.t.treasury) || {};
        },

        // Enter town
        enterTown() {
            this.g.sendCommand('select', '11');
        },

        // Town actions delegate to server options
        selectOption(key) {
            this.g.sendCommand('select', key);
        },

        // Prompt submission
        submitPrompt(value) {
            this.g.sendCommand('input', value);
        },

        barPct(current, max) {
            return this.g.barPct(current, max);
        }
    };
}
