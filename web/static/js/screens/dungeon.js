// Dungeon screen Alpine component
function dungeonScreen() {
    return {
        get g() { return this.$store.game; },
        get p() { return this.$store.game.player || {}; },
        get d() { return this.$store.game.dungeon; },

        get isDungeonScreen() {
            const s = this.g.serverScreen;
            return s && s.startsWith('dungeon_');
        },

        get hasActiveDungeon() {
            return this.d !== null;
        },

        get floor() {
            return this.d && this.d.floor ? this.d.floor : null;
        },

        get grid() {
            return this.floor && this.floor.grid ? this.floor.grid : null;
        },

        get isGridView() {
            const s = this.g.serverScreen;
            return s === 'dungeon_floor_map' || s === 'dungeon_grid_move';
        },

        get isRoomView() {
            const s = this.g.serverScreen;
            return s === 'dungeon_room' || s === 'dungeon_treasure' || s === 'dungeon_trap';
        },

        enterDungeon() {
            this.g.sendCommand('select', '12');
        },

        selectOption(key) {
            this.g.sendCommand('select', key);
        },

        // Grid tile rendering
        tileClass(tile, x, y) {
            if (!this.floor) return 'tile-wall';
            const isPlayer = x === this.floor.player_x && y === this.floor.player_y;
            const isExit = x === this.floor.exit_x && y === this.floor.exit_y;

            if (isPlayer) return 'tile-player';
            if (!tile.explored) return 'tile-fog';
            if (isExit) return 'tile-exit';

            switch (tile.type) {
                case 'room':
                    if (tile.cleared) return 'tile-room-cleared';
                    return 'tile-room';
                case 'corridor': return 'tile-corridor';
                case 'entrance': return 'tile-entrance';
                case 'exit': return 'tile-exit';
                default: return 'tile-wall';
            }
        },

        tileChar(tile, x, y) {
            if (!this.floor) return '';
            const isPlayer = x === this.floor.player_x && y === this.floor.player_y;
            const isExit = x === this.floor.exit_x && y === this.floor.exit_y;

            if (isPlayer) return '@';
            if (!tile.explored) return '';
            if (isExit) return 'E';

            switch (tile.type) {
                case 'room':
                    if (tile.room_type === 'monster') return tile.cleared ? '.' : '!';
                    if (tile.room_type === 'treasure') return tile.cleared ? '.' : '$';
                    if (tile.room_type === 'trap') return tile.cleared ? '.' : '^';
                    if (tile.room_type === 'rest') return tile.cleared ? '.' : '+';
                    if (tile.room_type === 'merchant') return tile.cleared ? '.' : 'M';
                    if (tile.room_type === 'boss') return tile.cleared ? '.' : 'B';
                    if (tile.room_type === 'investigation') return tile.cleared ? '.' : '?';
                    return tile.cleared ? '.' : '?';
                case 'corridor': return '.';
                case 'entrance': return 'S';
                default: return '';
            }
        },

        // Rooms cleared progress
        get roomsCleared() {
            if (!this.floor || !this.floor.rooms) return 0;
            return this.floor.rooms.filter(r => r.cleared).length;
        },

        get totalRooms() {
            return this.floor ? this.floor.total_rooms : 0;
        },

        get floorLabel() {
            if (!this.d) return '';
            return this.d.name + ' - Floor ' + (this.d.current_floor + 1) + '/' + this.d.total_floors;
        }
    };
}
