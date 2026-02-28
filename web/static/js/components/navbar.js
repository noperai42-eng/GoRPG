// Navigation bar + player resource bar component
function navbar() {
    return {
        tabs: [
            { id: 'hub', label: 'Hub', icon: '' },
            { id: 'map', label: 'Map', icon: '' },
            { id: 'village', label: 'Village', icon: '' },
            { id: 'town', label: 'Town', icon: '' },
            { id: 'dungeon', label: 'Dungeon', icon: '' },
            { id: 'quests', label: 'Quests', icon: '' },
            { id: 'stats', label: 'Stats', icon: '' },
        ],

        switchTab(tabId) {
            if (this.$store.game.inCombat) return;
            const ss = this.$store.game.serverScreen;
            // If leaving village or town, atomically return to main menu
            if (tabId !== 'village' && ss && ss.startsWith('village_')) {
                this.$store.game.sendCommand('select', 'home');
            } else if (tabId !== 'town' && ss && ss.startsWith('town_')) {
                this.$store.game.sendCommand('select', 'home');
            }
            this.$store.game.activeTab = tabId;
        },

        disconnect() {
            GameConnection.disconnect();
            this.$store.game.screen = 'characters';
            this.$store.game.player = null;
            this.$store.game.combat = null;
            this.$store.game.village = null;
            this.$store.game.town = null;
            this.$store.game.activeTab = 'hub';
        }
    };
}
