// Activity panel component (replaces former toast system)
function activityPanel() {
    return {
        expanded: true,
        categoryIcon(cat) {
            const icons = { combat: '\u2694', loot: '\uD83D\uDCE6', levelup: '\u2B06', heal: '\uD83D\uDC9A',
                damage: '\uD83D\uDCA5', buff: '\u2728', debuff: '\uD83D\uDD3B', narrative: '\uD83D\uDCDC',
                system: '\u2699', error: '\u26A0' };
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
