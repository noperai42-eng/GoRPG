// Alpine.js Global Store - central game state
document.addEventListener('alpine:init', () => {
    Alpine.store('game', {
        screen: 'auth',          // 'auth' | 'characters' | 'game'
        activeTab: 'hub',        // 'hub' | 'map' | 'village' | 'town' | 'quests' | 'stats'
        player: null,            // PlayerState from server
        combat: null,            // CombatView from server
        village: null,           // VillageView from server
        town: null,              // TownView from server
        dungeon: null,           // DungeonView from server
        serverScreen: null,      // raw screen field from server
        options: [],             // current server options
        prompt: null,            // current server prompt
        combatLog: [],           // messages during combat
        recentMessages: [],      // event groups: [{id, timestamp, messages, collapsed, isNew}]
        _groupId: 0,
        onlinePlayers: [],       // [{name, level, activity}]
        pendingAction: null,     // for two-step command chains
        dropdown: null,          // 'items' | 'skills' | null - combat dropdown
        showCompletedQuests: false,
        autoHunting: false,      // true when auto-hunting through multiple fights
        _autoHuntTimer: null,    // timer ID for auto-hunt delay
        version: '',             // server version string
        leaderboard: null,       // leaderboard entries from API
        leaderboardCategory: 'kills', // current leaderboard category
        mostWanted: null,        // most wanted monster entries from API

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

            // WASD keyboard support for dungeon grid movement
            document.addEventListener('keydown', (e) => {
                const s = Alpine.store('game').serverScreen;
                if (s !== 'dungeon_floor_map' && s !== 'dungeon_grid_move') return;
                // Don't intercept if user is typing in an input
                if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

                const keyMap = {
                    'w': 'n', 'W': 'n', 'ArrowUp': 'n',
                    's': 's', 'S': 's', 'ArrowDown': 's',
                    'a': 'w', 'A': 'w', 'ArrowLeft': 'w',
                    'd': 'e', 'D': 'e', 'ArrowRight': 'e',
                };
                const dir = keyMap[e.key];
                if (dir) {
                    e.preventDefault();
                    // Check if direction is available in current options
                    const opts = Alpine.store('game').options;
                    if (opts.some(o => o.key === dir)) {
                        Alpine.store('game').sendCommand('select', dir);
                    }
                }
            });
        },

        handleResponse(resp) {
            const isHarvestPush = resp.type === 'harvest';
            const isPresencePush = resp.type === 'presence';

            // Handle presence updates
            if (resp.state && resp.state.online_players) {
                this.onlinePlayers = resp.state.online_players;
            }
            if (isPresencePush) return;

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
                if (resp.state.dungeon) {
                    this.dungeon = resp.state.dungeon;
                } else if (!this.combat && resp.state.screen && !resp.state.screen.startsWith('dungeon')) {
                    this.dungeon = null;
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
                    // Outside combat, messages go to grouped event feed
                    const batchMsgs = [];
                    for (const msg of resp.messages) {
                        if (msg.text && msg.text.trim()) {
                            batchMsgs.push({ text: msg.text, category: msg.category || 'system' });
                        }
                    }
                    if (batchMsgs.length > 0) {
                        const gid = ++this._groupId;
                        this.recentMessages.push({
                            id: gid,
                            timestamp: Date.now(),
                            messages: batchMsgs,
                            collapsed: true,
                            isNew: true
                        });
                        setTimeout(() => {
                            const g = this.recentMessages.find(g => g.id === gid);
                            if (g) g.isNew = false;
                        }, 2000);
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
                setTimeout(() => {
                    GameConnection.disconnect();
                    this.screen = 'characters';
                    this.player = null;
                    this.combat = null;
                    this.village = null;
                    this.town = null;
                    this.activeTab = 'hub';
                    this.onlinePlayers = [];
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

            // Bounty board screens
            if (s === 'most_wanted_board' || s === 'most_wanted_hunt') {
                this.activeTab = 'hub';
                return;
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

            // Dungeon screens
            if (s === 'dungeon_select' || s === 'dungeon_floor_map' || s === 'dungeon_room' || s === 'dungeon_grid_move') {
                this.activeTab = 'dungeon';
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

        // Fetch leaderboard data from REST API
        fetchLeaderboard(category) {
            this.leaderboardCategory = category || this.leaderboardCategory;
            fetch('/api/leaderboard?category=' + encodeURIComponent(this.leaderboardCategory) + '&limit=20', {
                headers: Auth.getAuthHeaders()
            })
            .then(r => r.json())
            .then(data => {
                this.leaderboard = data.entries || [];
            })
            .catch(() => { this.leaderboard = []; });
        },

        // Fetch most wanted monsters from REST API
        fetchMostWanted() {
            fetch('/api/mostwanted?limit=10', {
                headers: Auth.getAuthHeaders()
            })
            .then(r => r.json())
            .then(data => {
                this.mostWanted = data.entries || [];
            })
            .catch(() => { this.mostWanted = []; });
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
