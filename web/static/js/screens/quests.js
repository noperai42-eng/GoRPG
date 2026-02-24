// Quest log screen Alpine component
function questsScreen() {
    return {
        get g() { return this.$store.game; },
        get p() { return this.$store.game.player; },

        get activeQuests() {
            if (!this.p || !this.p.active_quests) return [];
            return this.p.active_quests;
        },

        get completedQuests() {
            if (!this.p || !this.p.completed_quests) return [];
            return this.p.completed_quests;
        },

        get showCompleted() {
            return this.g.showCompletedQuests;
        },

        toggleCompleted() {
            this.g.showCompletedQuests = !this.g.showCompletedQuests;
        },

        questProgress(quest) {
            if (!quest.target || quest.target <= 0) return '0%';
            return Math.min(100, (quest.progress / quest.target) * 100) + '%';
        },

        questProgressText(quest) {
            return quest.progress + ' / ' + quest.target;
        },

        // Load quest data from server
        refreshQuests() {
            this.g.sendCommand('select', '9');
        }
    };
}
