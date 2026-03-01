// Map screen Alpine component
function mapScreen() {
    return {
        get p() { return this.$store.game.player || {}; },

        // Starter locations never had guardians
        _starterLocations: ['Home', 'Training Hall', 'Forest', 'Lake', 'Hills'],

        get allLocations() {
            if (!this.p) return [];
            const starters = this._starterLocations;
            const known = (this.p.known_locations || []).map(l => ({
                ...l,
                locked: false,
                guardianDefeated: !starters.includes(l.name)
            }));
            const locked = (this.p.locked_locations || []).map(l => ({
                ...l,
                locked: true,
                guardianDefeated: false
            }));
            return [...known, ...locked];
        },

        get showHuntModal() {
            const s = this.$store.game.serverScreen;
            return s === 'hunt_count_select' || s === 'hunt_tracking';
        },

        typeIcon(type) {
            const icons = {
                'Mix': '\u2694',     // crossed swords
                'Ruin': '\u2620',    // skull
                'Resource': '\u26CF', // pick
                'Trade': '\u2696',   // scales
                'Base': '\u2302',    // house
            };
            return icons[type] || '\u2022';
        },

        typeClass(type) {
            return 'type-' + (type || 'base').toLowerCase();
        },

        clickLocation(loc) {
            if (loc.type === 'Base') return;
            const alreadyOnHunt = this.$store.game.serverScreen === 'hunt_location_select';

            if (loc.locked) {
                const value = 'locked:' + loc.name;
                if (alreadyOnHunt) {
                    // Already on hunt select — send location directly
                    this.$store.game.sendCommand('select', value);
                } else {
                    // Need to navigate to hunt first, then auto-send location
                    this.$store.game.pendingAction = {
                        type: 'select',
                        value: value,
                        expectScreen: 'hunt_location_select'
                    };
                    this.$store.game.sendCommand('select', '3');
                }
            } else {
                if (alreadyOnHunt) {
                    // Already on hunt select — send location directly
                    this.$store.game.sendCommand('select', loc.name);
                } else {
                    // Need to navigate to hunt first, then auto-send location
                    this.$store.game.pendingAction = {
                        type: 'select',
                        value: loc.name,
                        expectScreen: 'hunt_location_select'
                    };
                    this.$store.game.sendCommand('select', '3');
                }
            }
        },

        // For hunt modal - submit hunt count
        huntCount: '1',
        submitHuntCount() {
            if (this.huntCount && parseInt(this.huntCount) > 0) {
                this.$store.game.sendCommand('input', this.huntCount);
            }
        },

        // For tracking select
        selectTarget(key) {
            this.$store.game.sendCommand('select', key);
        }
    };
}
