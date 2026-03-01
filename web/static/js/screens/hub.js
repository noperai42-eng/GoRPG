// Hub screen Alpine component
function hubScreen() {
    return {
        equipSlots: ['Head', 'Chest', 'Legs', 'Feet', 'Hands', 'Main Hand', 'Off Hand', 'Accessory'],
        selectedSlot: null,

        get p() { return this.$store.game.player || {}; },
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
            this.$store.game.sendCommand('select', 'hunt');
        },
        goHarvest() {
            this.$store.game.sendCommand('select', 'harvest');
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

        // Activity feed helpers
        categoryIcon(category) {
            const icons = {
                system: '\u2699', combat: '\u2694', damage: '\uD83D\uDCA5',
                heal: '\u2764', loot: '\uD83C\uDF81', buff: '\u2B06',
                debuff: '\u2B07', narrative: '\uD83D\uDCDC', error: '\u26A0',
                levelup: '\u2B50',
                broadcast: '\uD83D\uDCE2'
            };
            return icons[category] || '\u2699';
        },

        relativeTime(timestamp) {
            const diff = Math.floor((Date.now() - timestamp) / 1000);
            if (diff < 5) return 'just now';
            if (diff < 60) return diff + 's ago';
            const mins = Math.floor(diff / 60);
            if (mins < 60) return mins + 'm ago';
            const hrs = Math.floor(mins / 60);
            return hrs + 'h ago';
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
