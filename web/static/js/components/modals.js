// Modal system for prompts, item/skill pickers, and harvest
function modalSystem() {
    return {
        get showPromptModal() {
            const g = this.$store.game;
            return g.prompt && !g.inCombat;
        },

        get showOptionsModal() {
            const g = this.$store.game;
            // Show options as modal when we have options but no prompt,
            // and we're on a screen that needs modal-style display
            const modalScreens = ['harvest_select', 'hunt_count_select', 'hunt_tracking',
                'combat_guard_prompt', 'combat_skill_reward',
                'autoplay_speed', 'load_save_char_select',
                'character_select', 'character_create'];
            return !g.inCombat && !g.prompt && g.options.length > 0 && modalScreens.includes(g.serverScreen);
        },

        promptValue: '',

        submitPrompt() {
            if (this.promptValue.trim()) {
                this.$store.game.sendCommand('input', this.promptValue.trim());
                this.promptValue = '';
            }
        },

        selectOption(key) {
            this.$store.game.sendCommand('select', key);
        },

        getModalTitle() {
            const s = this.$store.game.serverScreen;
            const titles = {
                'harvest_select': 'Harvest Resources',
                'hunt_count_select': 'How Many Hunts?',
                'hunt_tracking': 'Select Target',
                'combat_guard_prompt': 'Bring Guards?',
                'combat_skill_reward': 'Skill Reward',
                'autoplay_speed': 'Auto-Play Speed',
                'load_save_char_select': 'Select Character',
                'character_select': 'Select Character',
                'character_create': 'Create Character',
            };
            return titles[s] || 'Choose an Option';
        }
    };
}
