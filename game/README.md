# Game Service

Game service for Gordle - handles single-player sessions and multiplayer rooms.

## Multiplayer Architecture

### REST API (via Gateway)

| Method | Endpoint                    | Description                     | Auth     |
|--------|-----------------------------|---------------------------------|----------|
| POST   | `/api/v1/rooms`             | Create room                     | Required |
| GET    | `/api/v1/rooms/{code}`      | Get room state                  | -        |
| POST   | `/api/v1/rooms/{code}/join` | Join room (returns `player_id`) | -        |

### WebSocket

Connect to: `wss://game-server/rooms/{code}/ws?player_id=xxx`

The `player_id` is obtained from `JoinRoom` response. Once connected, all game actions go through WebSocket.

#### Client Messages

```json
{"type": "ready", "payload": {"ready": true}}
{"type": "start_game"}
{"type": "guess", "payload": {"word": "crane"}}
{"type": "next_round"}
{"type": "leave"}
```

#### Server Events

| Event           | Description                         |
|-----------------|-------------------------------------|
| `player_joined` | A player joined the room            |
| `player_left`   | A player left the room              |
| `player_ready`  | A player changed ready status       |
| `game_started`  | Game/round started                  |
| `player_guess`  | A player made a guess               |
| `round_ended`   | Round finished                      |
| `game_ended`    | Game finished                       |
| `room_updated`  | Room state changed (e.g., new host) |
| `error`         | Error response                      |

#### Event Payloads

```json
// player_joined
{"player": {"player_id": "...", "display_name": "...", ...}}

// player_left
{"player_id": "...", "display_name": "..."}

// player_ready
{"player_id": "...", "display_name": "...", "ready": true}

// game_started
{"round_number": 1, "word_length": 5}

// player_guess
{"player_id": "...", "display_name": "...", "attempts": 3, "solved": false}

// round_ended
{"round_number": 1, "target_word": "crane", "results": [...]}

// game_ended
{"reason": "round_complete", "final_scores": [...], "target_word": "crane"}

// error
{"code": "guess_failed", "message": "word is not in dictionary"}
```

## Scaling (Redis Pub/Sub)

WebSocket connections are distributed across multiple game service instances using Redis Pub/Sub.

```
Instance 1                         Redis                        Instance 2
┌─────────┐                    ┌─────────┐                    ┌─────────┐
│   Hub   │                    │ Channel │                    │   Hub   │
│ ┌─────┐ │  publish           │         │  subscribe         │ ┌─────┐ │
│ │ A   │───────────────────────▶ws:room:X─────────────────────▶│ B   │ │
│ └─────┘ │                    │         │                    │ └─────┘ │
└─────────┘                    └─────────┘                    └─────────┘
```

**How it works:**
- First local player joins a room → Hub subscribes to `ws:room:{code}` channel
- Broadcast event → publish to Redis → all instances receive → deliver to local connections
- Last local player leaves → Hub unsubscribes from channel

**Channel format:** `ws:room:{roomCode}`

**Local-only delivery:** Error messages (`SendToPlayer`) are delivered locally only, not through Redis.

## Flow Example

```
1. Host: POST /api/v1/rooms → {room: {code: "ABC123"}, ...}
2. Host: WS connect /rooms/ABC123/ws?player_id=host-uuid

3. Guest: POST /api/v1/rooms/ABC123/join {guest_id: "xxx", display_name: "Player2"}
        → {player: {player_id: "guest:xxx"}, ...}
4. Guest: WS connect /rooms/ABC123/ws?player_id=guest:xxx

5. Guest: → {"type": "ready", "payload": {"ready": true}}
   All:   ← {"type": "player_ready", ...}

6. Host:  → {"type": "start_game"}
   All:   ← {"type": "game_started", "payload": {"round_number": 1, "word_length": 5}}

7. Guest: → {"type": "guess", "payload": {"word": "crane"}}
   All:   ← {"type": "player_guess", ...}

8. (game continues until all finish)
   All:   ← {"type": "round_ended", ...}
   All:   ← {"type": "game_ended", ...}
```
