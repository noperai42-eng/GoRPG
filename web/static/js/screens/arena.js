// Arena screen Alpine component
function arenaScreen() {
    return {
        get g() { return this.$store.game; },
        get p() { return this.$store.game.player; },

        init() {
            this.g.fetchArena();
        },

        get arenaData() {
            return this.g.arena || { entries: [], champion: null };
        },

        get arenaEntries() {
            return this.arenaData.entries || [];
        },

        get arenaChampion() {
            return this.arenaData.champion || null;
        },

        refreshArena() {
            this.g.fetchArena();
        },

        arenaRankClass(idx) {
            if (idx === 0) return 'rank-gold';
            if (idx === 1) return 'rank-silver';
            if (idx === 2) return 'rank-bronze';
            return '';
        },

        isArenaSelf(entry) {
            return this.p && entry.character_name === this.p.name;
        },

        enterArena() {
            this.g.sendCommand('select', '14');
        },

        challengePlayer(entry) {
            this.g.sendCommand('select', 'arena_challenge:' + entry.account_id + ':' + entry.character_name);
        }
    };
}
