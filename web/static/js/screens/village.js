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
        _harvestTimer: null,
        _harvestTick: 0,

        get g() { return this.$store.game; },
        get v() { return this.$store.game.village; },
        get p() { return this.$store.game.player; },
        get hasVillage() { return this.v !== null; },

        get isVillageScreen() {
            const s = this.g.serverScreen;
            return s && s.startsWith('village_');
        },

        init() {
            // Refresh countdown every second
            this._harvestTimer = setInterval(() => { this._harvestTick++; }, 1000);
        },

        destroy() {
            if (this._harvestTimer) {
                clearInterval(this._harvestTimer);
                this._harvestTimer = null;
            }
        },

        // Seconds until next harvest (60s cycle)
        harvestCountdown() {
            // Reference _harvestTick to trigger reactivity
            void this._harvestTick;
            if (!this.v || !this.v.last_harvest_time) return 60;
            const elapsed = Math.floor(Date.now() / 1000) - this.v.last_harvest_time;
            const remaining = 60 - elapsed;
            return remaining > 0 ? remaining : 0;
        },

        // List of villagers actively harvesting
        activeHarvesters() {
            if (!this.v || !this.v.villagers) return [];
            return this.v.villagers.filter(vil => vil.role === 'harvester' && vil.harvest_type);
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
