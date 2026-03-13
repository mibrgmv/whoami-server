# Gordle Web

React frontend for Gordle - multiplayer word game.

## Stack

- **React 18** + TypeScript
- **Vite** - build tool
- **Zustand** - state management
- **React Router** - routing

## Project Structure

```
src/
├── api/
│   └── client.ts         # API client (auth, game, room)
├── components/
│   ├── Tile.tsx          # Single letter tile
│   ├── GameBoard.tsx     # 5x6 grid
│   └── Keyboard.tsx      # Virtual keyboard
├── pages/
│   ├── Home.tsx          # Landing (solo/create/join)
│   ├── Game.tsx          # Solo game
│   └── Room.tsx          # Multiplayer (lobby + game)
├── stores/
│   ├── authStore.ts      # Auth state (tokens, guest)
│   ├── gameStore.ts      # Solo game state
│   └── roomStore.ts      # Multiplayer + WebSocket
├── types/
│   ├── api.ts            # API types (from proto)
│   └── ws.ts             # WebSocket event types
└── App.tsx               # Router setup
```

## Development

```bash
# Install dependencies
npm install

# Start dev server (port 3000)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Features

- Solo game (random word)
- Multiplayer rooms via WebSocket
- Guest authentication (auto-login)
- Physical + virtual keyboard support
- Tile flip animations
- Dark theme (Wordle-style)

## API Proxy

Dev server proxies `/api/*` to `http://localhost:8080` (gateway).

## TODO

- [ ] Login/Register pages (currently guest-only)
- [ ] Daily challenge mode
- [ ] Russian keyboard layout
- [ ] Sound effects
- [ ] More animations
- [ ] Statistics page
- [ ] Profile page
