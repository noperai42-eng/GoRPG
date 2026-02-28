// Stats & Leaderboard screen Alpine component
function statsScreen() {
    return {
        subTab: 'mystats', // 'mystats' | 'leaderboard' | 'mostwanted'
        _leaderboardLoaded: false,
        _mostWantedLoaded: false,

        get g() { return this.$store.game; },
        get p() { return this.$store.game.player; },
        get stats() { return this.p ? this.p.stats : null; },

        init() {
            this.$watch('subTab', (val) => {
                if (val === 'leaderboard' && !this._leaderboardLoaded) {
                    this._leaderboardLoaded = true;
                    this.g.fetchLeaderboard();
                }
                if (val === 'mostwanted' && !this._mostWantedLoaded) {
                    this._mostWantedLoaded = true;
                    this.g.fetchMostWanted();
                }
            });
        },

        // K/D ratio
        kdRatio() {
            if (!this.stats) return '0.00';
            if (this.stats.total_deaths === 0) return this.stats.total_kills.toFixed(2);
            return (this.stats.total_kills / this.stats.total_deaths).toFixed(2);
        },

        // PvP win rate
        pvpWinRate() {
            if (!this.stats) return '0%';
            const total = this.stats.pvp_wins + this.stats.pvp_losses;
            if (total === 0) return '0%';
            return Math.round((this.stats.pvp_wins / total) * 100) + '%';
        },

        // Kills by rarity entries
        rarityEntries() {
            if (!this.stats || !this.stats.kills_by_rarity) return [];
            return Object.entries(this.stats.kills_by_rarity).sort((a, b) => b[1] - a[1]);
        },

        // Leaderboard helpers
        switchCategory(cat) {
            this.g.fetchLeaderboard(cat);
        },

        get leaderboardEntries() {
            return this.g.leaderboard || [];
        },

        get currentCategory() {
            return this.g.leaderboardCategory;
        },

        categoryLabel(cat) {
            const labels = { kills: 'Kills', level: 'Level', bosses: 'Bosses', pvp_wins: 'PvP', combo: 'Combo', dungeons: 'Dungeons', floors: 'Floors', rooms: 'Rooms' };
            return labels[cat] || cat;
        },

        categoryValue(entry, cat) {
            switch (cat) {
                case 'kills': return entry.total_kills;
                case 'level': return entry.player_level;
                case 'bosses': return entry.bosses_killed;
                case 'pvp_wins': return entry.pvp_wins;
                case 'combo': return entry.highest_combo;
                case 'dungeons': return entry.dungeons_cleared;
                case 'floors': return entry.floors_cleared;
                case 'rooms': return entry.rooms_explored;
                default: return entry.total_kills;
            }
        },

        isSelf(entry) {
            return this.p && entry.character_name === this.p.name;
        },

        rankClass(idx) {
            if (idx === 0) return 'rank-gold';
            if (idx === 1) return 'rank-silver';
            if (idx === 2) return 'rank-bronze';
            return '';
        },

        // Most Wanted helpers
        get mostWantedEntries() {
            return this.g.mostWanted || [];
        },

        refreshMostWanted() {
            this.g.fetchMostWanted();
        },

        rarityClass(rarity) {
            if (!rarity) return 'rarity-common';
            return 'rarity-' + rarity;
        },

        rarityLabel(rarity) {
            if (!rarity || rarity === 'common') return 'Common';
            const labels = { uncommon: 'Uncommon', rare: 'Rare', epic: 'Epic', legendary: 'Legendary', mythic: 'Mythic' };
            return labels[rarity] || 'Common';
        },

        totalKills(entry) {
            return entry.player_kills || 0;
        }
    };
}
