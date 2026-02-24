// Combat screen Alpine component
function combatScreen() {
    return {
        get g() { return this.$store.game; },
        get c() { return this.$store.game.combat; },
        get p() { return this.$store.game.player; },

        get playerInitials() {
            if (!this.p) return '?';
            return this.p.name.substring(0, 2).toUpperCase();
        },

        get monsterInitials() {
            if (!this.c) return '?';
            return this.c.monster_name.substring(0, 2).toUpperCase();
        },

        get monsterPanelClass() {
            if (!this.c) return '';
            if (this.c.monster_is_boss) return 'boss';
            if (this.c.monster_is_guardian) return 'guardian';
            return '';
        },

        get avatarClass() {
            if (!this.c) return 'monster';
            if (this.c.monster_is_boss) return 'boss';
            return 'monster';
        },

        get hasGuards() {
            return this.c && this.c.guards && this.c.guards.length > 0;
        },

        get isItemSelect() {
            return this.g.serverScreen === 'combat_item_select';
        },

        get isSkillSelect() {
            return this.g.serverScreen === 'combat_skill_select';
        },

        get showDropdown() {
            return this.g.dropdown !== null;
        },

        // Combat actions
        doAttack() {
            this.g.dropdown = null;
            this.g.sendCommand('select', '1');
        },
        doDefend() {
            this.g.dropdown = null;
            this.g.sendCommand('select', '2');
        },
        doItem() {
            if (this.g.dropdown === 'items') {
                this.g.dropdown = null;
                return;
            }
            this.g.dropdown = 'items';
            this.g.sendCommand('select', '3');
        },
        doSkill() {
            if (this.g.dropdown === 'skills') {
                this.g.dropdown = null;
                return;
            }
            this.g.dropdown = 'skills';
            this.g.sendCommand('select', '4');
        },
        doFlee() {
            this.g.dropdown = null;
            this.g.sendCommand('select', '5');
        },
        doAuto() {
            this.g.dropdown = null;
            this.g.sendCommand('select', '6');
        },

        closeDropdown() {
            this.g.dropdown = null;
        },

        selectDropdownItem(key) {
            this.g.dropdown = null;
            this.g.sendCommand('select', key);
        },

        // Options for dropdown (from server)
        get dropdownOptions() {
            if (this.isItemSelect || this.isSkillSelect) {
                return this.g.options;
            }
            return [];
        },

        get dropdownTitle() {
            if (this.isItemSelect) return 'Use Item';
            if (this.isSkillSelect) return 'Use Skill';
            return '';
        },

        // Guard prompt / skill reward
        get showGuardPrompt() {
            return this.g.serverScreen === 'combat_guard_prompt';
        },

        get showSkillReward() {
            return this.g.serverScreen === 'combat_skill_reward';
        },

        // For auto-scroll of combat log
        scrollLog() {
            this.$nextTick(() => {
                const el = this.$refs.combatLog;
                if (el) el.scrollTop = el.scrollHeight;
            });
        },

        // Helper for bar width
        barPct(current, max) {
            return this.$store.game.barPct(current, max);
        },

        hpClass(current, max) {
            return this.$store.game.hpClass(current, max);
        }
    };
}
