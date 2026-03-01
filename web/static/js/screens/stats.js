// Stats & Leaderboard screen Alpine component
function statsScreen() {
    return {
        subTab: 'mystats', // 'mystats' | 'leaderboard' | 'mostwanted'

        get g() { return this.$store.game; },
        get p() { return this.$store.game.player || {}; },
        get stats() { return (this.p && this.p.stats) ? this.p.stats : null; },

        // Called when the stats tab opens (x-init)
        loadData() {
            // Auto-fetch leaderboard on tab open
            setTimeout(() => {
                this.g.fetchLeaderboard();
                this.g.fetchMostWanted();
            }, 100);
        },

        switchToLeaderboard() {
            this.subTab = 'leaderboard';
            this.g.fetchLeaderboard();
        },

        switchToMostWanted() {
            this.subTab = 'mostwanted';
            this.g.fetchMostWanted();
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

        // Render kills by rarity as HTML (avoids x-for)
        renderRarityTable() {
            if (!this.stats || !this.stats.kills_by_rarity) return '';
            const entries = Object.entries(this.stats.kills_by_rarity).sort((a, b) => b[1] - a[1]);
            if (entries.length === 0) return '';
            let html = '<div class="section-header">Kills by Rarity</div>';
            for (const [rarity, count] of entries) {
                html += `<div class="stat-row"><span class="stat-label">${this._esc(rarity)}</span><span class="stat-value">${count}</span></div>`;
            }
            return html;
        },

        // Render online players as HTML (avoids x-for)
        renderOnlinePlayers() {
            const players = this.g.onlinePlayers;
            if (!players || players.length === 0) return '';
            let html = '';
            for (const op of players) {
                html += `<div class="stat-row"><span class="stat-label">${this._esc(op.name)} (Lv ${op.level})</span><span class="stat-value">${this._esc(op.activity)}</span></div>`;
            }
            return html;
        },

        // Render leaderboard as HTML (avoids x-for)
        renderLeaderboard() {
            const entries = this.g.leaderboard || [];
            if (entries.length === 0) return '';
            const cat = this.g.leaderboardCategory;
            const selfName = this.p ? this.p.name : '';
            let html = '';
            entries.forEach((entry, idx) => {
                const isSelf = entry.character_name === selfName;
                const rankCls = idx === 0 ? 'rank-gold' : idx === 1 ? 'rank-silver' : idx === 2 ? 'rank-bronze' : '';
                const selfCls = isSelf ? ' leaderboard-self' : '';
                const val = this.categoryValue(entry, cat);
                html += `<div class="leaderboard-row${selfCls}">`;
                html += `<span class="leaderboard-rank ${rankCls}">#${idx + 1}</span>`;
                html += `<span class="leaderboard-name">${this._esc(entry.character_name)}</span>`;
                html += `<span class="leaderboard-level">Lv ${entry.player_level}</span>`;
                html += `<span class="leaderboard-score">${val}</span>`;
                html += '</div>';
            });
            return html;
        },

        // Render most wanted as HTML (avoids x-for)
        renderMostWanted() {
            const entries = this.g.mostWanted || [];
            if (entries.length === 0) return '';
            let html = '';
            entries.forEach((entry, idx) => {
                const rankCls = idx === 0 ? 'rank-gold' : idx === 1 ? 'rank-silver' : idx === 2 ? 'rank-bronze' : '';
                const rarCls = entry.rarity ? 'rarity-' + entry.rarity : 'rarity-common';
                const rarLbl = this.rarityLabel(entry.rarity);
                html += '<div class="leaderboard-row">';
                html += `<span class="leaderboard-rank ${rankCls}">#${idx + 1}</span>`;
                html += `<span class="most-wanted-rarity-badge ${rarCls}">${this._esc(rarLbl)}</span>`;
                html += `<span class="leaderboard-name">${this._esc(entry.name)}</span>`;
                html += `<span class="leaderboard-level">Lv ${entry.level}</span>`;
                html += `<span class="most-wanted-location">${this._esc(entry.location_name)}</span>`;
                html += `<span class="leaderboard-score">${entry.player_kills || 0} player kills</span>`;
                html += '</div>';
            });
            return html;
        },

        // HTML-escape helper
        _esc(s) {
            if (!s) return '';
            return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
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
