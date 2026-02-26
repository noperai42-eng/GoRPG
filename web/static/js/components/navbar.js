// Navigation bar + player resource bar component
function navbar() {
    return {
        tabs: [
            { id: 'hub', label: 'Hub', icon: '' },
            { id: 'map', label: 'Map', icon: '' },
            { id: 'village', label: 'Village', icon: '' },
            { id: 'town', label: 'Town', icon: '' },
            { id: 'quests', label: 'Quests', icon: '' },
        ],

        switchTab(tabId) {
            if (this.$store.game.inCombat) return;
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
