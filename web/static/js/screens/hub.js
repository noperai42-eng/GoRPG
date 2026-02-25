// Hub screen Alpine component
function hubScreen() {
    return {
        equipSlots: ['Head', 'Chest', 'Legs', 'Feet', 'Hands', 'Main Hand', 'Off Hand', 'Accessory'],
        selectedSlot: null,

        get p() { return this.$store.game.player; },
        get hasAutoplay() {
            const s = this.$store.game.serverScreen;
            return s && s.startsWith('autoplay_');
        },

        getEquipItem(slotName) {
            if (!this.p || !this.p.equipment) return null;
            return this.p.equipment[slotName] || null;
        },

        toggleItemInfo(slot) {
            if (this.selectedSlot === slot) {
                this.selectedSlot = null;
            } else if (this.getEquipItem(slot)) {
                this.selectedSlot = slot;
            }
        },

        rarityClass(item) {
            if (!item) return '';
            const r = item.rarity || 1;
            return 'rarity-' + Math.min(r, 5);
        },

        rarityLabel(item) {
            if (!item) return '';
            const labels = {1: 'Common', 2: 'Uncommon', 3: 'Rare', 4: 'Epic', 5: 'Legendary'};
            return labels[Math.min(item.rarity || 1, 5)] || 'Common';
        },

        skillCost(skill) {
            const parts = [];
            if (skill.mana_cost > 0) parts.push(skill.mana_cost + 'MP');
            if (skill.stamina_cost > 0) parts.push(skill.stamina_cost + 'SP');
            return parts.join(' ') || 'Free';
        },

        // Quick actions
        goHunt() {
            // Send hunt command to server to enter hunt_location_select state,
            // then switch to map tab (routing happens automatically via _routeScreen)
            if (this.$store.game.serverScreen === 'main_menu') {
                this.$store.game.sendCommand('select', '3');
            } else {
                // Already past main menu — just switch to map tab
                this.$store.game.activeTab = 'map';
            }
        },
        goHarvest() {
            this.$store.game.sendCommand('select', '1');
        },
        goAutoPlay() {
            this.$store.game.sendCommand('select', '8');
        },
        goGuide() {
            this.$store.game.sendCommand('select', '7');
        },
        goVillage() {
            this.$store.game.sendCommand('select', '10');
        },
        saveExit() {
            this.$store.game.sendCommand('select', 'exit');
        },

        // Handle autoplay menu options
        autoplayAction(key) {
            this.$store.game.sendCommand('select', key);
        },

        get isAutoPlayMenu() {
            return this.$store.game.serverScreen === 'autoplay_menu';
        },

        get autoplayOptions() {
            if (!this.isAutoPlayMenu) return [];
            return this.$store.game.options;
        }
    };
}
