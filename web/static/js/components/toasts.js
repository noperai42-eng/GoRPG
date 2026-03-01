// Activity panel component (replaces former toast system)
function activityPanel() {
    return {
        expanded: true,
        // Stable computed list — avoids creating new arrays in x-for expression
        get displayGroups() {
            const msgs = this.$store.game.recentMessages;
            if (!msgs || msgs.length === 0) return [];
            // Return last 10 in reverse order; slice+reverse on a plain array
            const start = Math.max(0, msgs.length - 10);
            const result = [];
            for (let i = msgs.length - 1; i >= start; i--) {
                result.push(msgs[i]);
            }
            return result;
        },
        // Single header message for a group (replaces nested x-if branching)
        headerMsg(group) {
            if (group.messages.length === 1) return group.messages[0];
            return this.$store.game.groupHeader(group);
        },
        categoryIcon(cat) {
            const icons = { combat: '\u2694', loot: '\uD83D\uDCE6', levelup: '\u2B06', heal: '\uD83D\uDC9A',
                damage: '\uD83D\uDCA5', buff: '\u2728', debuff: '\uD83D\uDD3B', narrative: '\uD83D\uDCDC',
                system: '\u2699', error: '\u26A0', broadcast: '\uD83D\uDCE2' };
            return icons[cat] || '\u2022';
        },
        relativeTime(ts) {
            const diff = Math.floor((Date.now() - ts) / 1000);
            if (diff < 60) return 'now';
            if (diff < 3600) return Math.floor(diff / 60) + 'm';
            return Math.floor(diff / 3600) + 'h';
        }
    };
}

// Keep toastSystem as a no-op for backward compat if still referenced
function toastSystem() { return {}; }
