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
            { id: 'arena', label: 'Arena', icon: '' },
        ],

        switchTab(tabId) {
            if (this.$store.game.inCombat) return;
            const ss = this.$store.game.serverScreen;
            const g = this.$store.game;

            g.activeTab = tabId;

            // Auto-navigate to server-driven tabs (village/town/dungeon)
            const entryCommands = { village: '10', town: '11', dungeon: '12' };
            const cmd = entryCommands[tabId];
            if (cmd) {
                // Don't auto-enter dungeon if already inside one (would lose progress)
                if (tabId === 'dungeon' && ss && ss.startsWith('dungeon_')) return;

                const hubScreens = ['main_menu', 'harvest_select', 'player_stats', 'discovered_locations'];
                if (hubScreens.includes(ss)) {
                    // On hub, directly enter
                    g.sendCommand('select', cmd);
                } else {
                    // Navigate home first, then auto-enter via pending action
                    g.sendCommand('select', 'home');
                    g.pendingAction = { expectScreen: 'main_menu', value: cmd };
                }
                return;
            }

            // For non-server tabs, return home if in village/town
            if (ss && ss.startsWith('village_')) {
                g.sendCommand('select', 'home');
            } else if (ss && ss.startsWith('town_')) {
                g.sendCommand('select', 'home');
            }
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
