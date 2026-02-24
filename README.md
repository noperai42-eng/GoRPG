# RPG Game

A text-based RPG with a web UI featuring tactical turn-based combat, village management, questing, and resource harvesting.

## Prerequisites

- **Go 1.24+** (uses SQLite via cgo, so a C compiler is also required)
- A modern web browser

## Building

```bash
go build -o rpg-server ./cmd/server/
```

This produces a single `rpg-server` binary.

## Running

```bash
./rpg-server
```

The server starts on `http://localhost:8080` by default. Open that URL in a browser to play.

### Command-line flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8080` | Listen address (host:port) |
| `-db` | `game.db` | Path to SQLite database file |
| `-secret` | `change-me-in-production` | JWT signing secret for auth tokens |
| `-static` | `web/static` | Path to the static files directory |

Example with custom settings:

```bash
./rpg-server -addr :3000 -db /var/data/rpg.db -secret my-secret-key
```

### Static files

The `-static` flag must point to the `web/static` directory (or a copy of it). When running from the project root, the default `web/static` works. When deploying the binary elsewhere, copy `web/static/` alongside it and set the flag appropriately:

```bash
# Deploy example
cp -r web/static /opt/rpg/static
cp rpg-server /opt/rpg/
cd /opt/rpg && ./rpg-server -static ./static
```

## Playing

1. Open `http://localhost:8080` in a browser
2. Register an account or log in
3. Create a character (or auto-connect to an existing one)
4. Use the tab-based UI: **Hub** (character stats, quick actions), **Map** (hunt locations), **Village** (management), **Quests** (quest log)
5. Combat takes over the full screen when you enter a fight

## Project Structure

```
cmd/server/          Server entrypoint
pkg/
  auth/              JWT authentication
  db/                SQLite database layer
  engine/            Game engine (state machine, combat, menus, village)
  game/              Game session management
  models/            Data models (Character, Monster, Item, Skill, etc.)
  server/            HTTP + WebSocket server
web/static/
  index.html         Alpine.js single-page app
  alpine.min.js      Alpine.js v3 library
  css/               Stylesheets (variables, layout, components, screens)
  js/
    auth.js           Auth (login/register/token management)
    websocket.js      WebSocket connection to game server
    store.js          Alpine.js global store (central game state)
    components/       Navbar, toasts, modals
    screens/          Hub, map, combat, village, quests
```

## Testing

```bash
go test ./...
```
