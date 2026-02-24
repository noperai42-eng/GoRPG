// Village screen Alpine component
function villageScreen() {
    return {
        subTab: 'overview',
        subTabs: [
            { id: 'overview', label: 'Overview' },
            { id: 'villagers', label: 'Villagers' },
            { id: 'guards', label: 'Guards' },
            { id: 'crafting', label: 'Crafting' },
            { id: 'defenses', label: 'Defenses' },
        ],

        get g() { return this.$store.game; },
        get v() { return this.$store.game.village; },
        get p() { return this.$store.game.player; },
        get hasVillage() { return this.v !== null; },

        get isVillageScreen() {
            const s = this.g.serverScreen;
            return s && s.startsWith('village_');
        },

        // Enter village
        enterVillage() {
            this.g.sendCommand('select', '10');
        },

        // Village actions delegate to server options
        selectOption(key) {
            this.g.sendCommand('select', key);
        },

        // Back to main menu from village
        backToMenu() {
            this.g.sendCommand('select', '0');
        },

        barPct(current, max) {
            return this.g.barPct(current, max);
        }
    };
}
