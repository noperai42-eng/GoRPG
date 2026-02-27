// Alpine.js Global Store - central game state
document.addEventListener('alpine:init', () => {
    Alpine.store('game', {
        screen: 'auth',          // 'auth' | 'characters' | 'game'
        activeTab: 'hub',        // 'hub' | 'map' | 'village' | 'town' | 'quests'
        player: null,            // PlayerState from server
        combat: null,            // CombatView from server
        village: null,           // VillageView from server
        town: null,              // TownView from server
        serverScreen: null,      // raw screen field from server
        options: [],             // current server options
        prompt: null,            // current server prompt
        toasts: [],              // [{text, category, id}]
        combatLog: [],           // messages during combat
        recentMessages: [],      // event groups: [{id, timestamp, messages, collapsed}]
        _groupId: 0,
        pendingAction: null,     // for two-step command chains
        dropdown: null,          // 'items' | 'skills' | null - combat dropdown
        showCompletedQuests: false,
        autoHunting: false,      // true when auto-hunting through multiple fights
        _autoHuntTimer: null,    // timer ID for auto-hunt delay
        _toastId: 0,
        version: '',             // server version string

        get inCombat() { return this.combat !== null; },

        init() {
            // Auto-login if token exists in storage
            Auth.init();
            if (Auth.isLoggedIn()) {
                this.screen = 'characters';
            }
            // Fetch server version
            fetch('/api/version').then(r => r.json()).then(d => {
                if (d.version) Alpine.store('game').version = 'v' + d.version;
            }).catch(() => {});
        },

        handleResponse(resp) {
            const isHarvestPush = resp.type === 'harvest';

            // Update player state
            if (resp.state) {
                if (resp.state.player) {
                    // Preserve enriched location data (type, level_max) if new response only has names
                    if (this.player) {
                        const oldKnown = this.player.known_locations;
                        const oldLocked = this.player.locked_locations;
                        const newP = resp.state.player;
                        if (oldKnown && newP.known_locations && newP.known_locations.length > 0 && !newP.known_locations[0].type) {
                            const locMap = {};
                            for (const l of oldKnown) if (l.type) locMap[l.name] = l;
                            newP.known_locations = newP.known_locations.map(l => locMap[l.name] || l);
                        }
                        if (oldLocked && newP.locked_locations && newP.locked_locations.length > 0 && !newP.locked_locations[0].type) {
                            const locMap = {};
                            for (const l of oldLocked) if (l.type) locMap[l.name] = l;
                            newP.locked_locations = newP.locked_locations.map(l => locMap[l.name] || l);
                        }
                    }
                    this.player = resp.state.player;
                }
                if (resp.state.combat) {
                    this.combat = resp.state.combat;
                } else if (this.combat !== null) {
                    // Combat ended
                    this.combat = null;
                    this.combatLog = [];
                    this.dropdown = null;
                }
                if (resp.state.village) {
                    this.village = resp.state.village;
                }
                if (resp.state.town) {
                    this.town = resp.state.town;
                }
                if (!isHarvestPush) {
                    this.serverScreen = resp.state.screen || null;
                }
            }

            // Update options (skip for harvest push to avoid clearing current menu)
            if (!isHarvestPush) {
                this.options = resp.options || [];
                this.prompt = resp.prompt || null;
            }

            // Route messages
            if (resp.messages && resp.messages.length > 0) {
                if (this.combat) {
                    // During combat, messages go to combat log
                    for (const msg of resp.messages) {
                        if (msg.text) {
                            this.combatLog.push(msg);
                        }
                    }
                    // Keep log manageable
                    if (this.combatLog.length > 100) {
                        this.combatLog = this.combatLog.slice(-80);
                    }
                } else {
                    // Outside combat, messages become toasts + grouped event feed
                    const batchMsgs = [];
                    for (const msg of resp.messages) {
                        if (msg.text && msg.text.trim()) {
                            this.addToast(msg.text, msg.category || 'system');
                            batchMsgs.push({ text: msg.text, category: msg.category || 'system' });
                        }
                    }
                    if (batchMsgs.length > 0) {
                        this.recentMessages.push({
                            id: ++this._groupId,
                            timestamp: Date.now(),
                            messages: batchMsgs,
                            collapsed: true
                        });
                        if (this.recentMessages.length > 15) {
                            this.recentMessages = this.recentMessages.slice(-15);
                        }
                    }
                }
            }

            // Handle pending action (auto-send for two-step chains)
            if (this.pendingAction && this.serverScreen) {
                const pa = this.pendingAction;
                if (pa.expectScreen && this.serverScreen === pa.expectScreen) {
                    this.pendingAction = null;
                    // Find matching option
                    const matchOpt = this.options.find(o => o.key === pa.value);
                    if (matchOpt) {
                        setTimeout(() => this.sendCommand('select', pa.value), 50);
                        return;
                    }
                }
                // Clear stale pending actions
                if (!pa.expectScreen) {
                    this.pendingAction = null;
                }
            }

            // Auto-hunt: if active and we're in combat, schedule next auto-fight
            if (this.autoHunting && this.combat && this.serverScreen === 'combat') {
                clearTimeout(this._autoHuntTimer);
                this._autoHuntTimer = setTimeout(() => {
                    if (this.autoHunting && this.combat) {
                        this.sendCommand('select', '6');
                    }
                }, 2000);
            }

            // Stop auto-hunting when combat ends
            if (!this.combat && this.autoHunting) {
                this.autoHunting = false;
                clearTimeout(this._autoHuntTimer);
            }

            // Screen routing based on serverScreen
            this._routeScreen();

            // Handle exit
            if (resp.type === 'exit') {
                this.addToast('Game saved. Disconnecting...', 'system');
                setTimeout(() => {
                    GameConnection.disconnect();
                    this.screen = 'characters';
                    this.player = null;
                    this.combat = null;
                    this.village = null;
                    this.town = null;
                    this.activeTab = 'hub';
                }, 1500);
            }
        },

        _routeScreen() {
            const s = this.serverScreen;
            if (!s) return;

            // Combat screens show combat overlay (handled reactively via inCombat)
            if (s === 'combat' || s === 'combat_item_select' || s === 'combat_skill_select' || s === 'combat_guard_prompt' || s === 'combat_skill_reward') {
                return; // combat overlay handles this
            }

            // Map screens
            if (s === 'hunt_location_select' || s === 'hunt_count_select' || s === 'hunt_tracking') {
                this.activeTab = 'map';
                return;
            }

            // Village screens
            if (s.startsWith('village_')) {
                this.activeTab = 'village';
                return;
            }

            // Town screens
            if (s.startsWith('town_')) {
                this.activeTab = 'town';
                return;
            }

            // Quest log
            if (s === 'quest_log') {
                this.activeTab = 'quests';
                return;
            }

            // Auto-play screens
            if (s.startsWith('autoplay_')) {
                this.activeTab = 'hub';
                return;
            }

            // Default - stay on current tab or go to hub
            if (s === 'main_menu' || s === 'harvest_select' || s === 'player_stats' || s === 'discovered_locations') {
                this.activeTab = 'hub';
            }
        },

        sendCommand(type, value) {
            GameConnection.sendCommand(type, value || '');
        },

        addToast(text, category) {
            const id = ++this._toastId;
            this.toasts.push({ text, category, id });
            setTimeout(() => this.removeToast(id), 4000);
            // Limit max visible toasts
            if (this.toasts.length > 8) {
                this.toasts = this.toasts.slice(-8);
            }
        },

        removeToast(id) {
            this.toasts = this.toasts.filter(t => t.id !== id);
        },

        // Helper: get HP bar class
        hpClass(current, max) {
            if (max <= 0) return 'hp';
            const pct = current / max;
            if (pct <= 0.1) return 'hp critical';
            if (pct <= 0.25) return 'hp low';
            if (pct <= 0.5) return 'hp mid';
            return 'hp';
        },

        // Helper: get bar width %
        barPct(current, max) {
            if (max <= 0) return '0%';
            return Math.max(0, Math.min(100, (current / max) * 100)) + '%';
        },

        // Toggle an event group's collapsed state
        toggleGroup(groupId) {
            const group = this.recentMessages.find(g => g.id === groupId);
            if (group) group.collapsed = !group.collapsed;
        },

        // Pick the most interesting message from a group as its header
        groupHeader(group) {
            const priority = ['levelup','loot','combat','heal','damage','buff','debuff','narrative','system','error'];
            let best = group.messages[0];
            let bestIdx = priority.length;
            for (const msg of group.messages) {
                const idx = priority.indexOf(msg.category);
                if (idx !== -1 && idx < bestIdx) {
                    best = msg;
                    bestIdx = idx;
                }
            }
            return best;
        },
    });
});
