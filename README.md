# Cinema Ticket Booking System

Online cinema ticket booking with concurrent seat locking, real-time updates, and admin dashboard.

See [cinema -ticket-flow.MD](./cinema%20-ticket-flow.MD) for the full implementation guide.

## Quick Start

```bash
docker compose up --build
```

## Tech Stack

- Backend: Go + Gin
- Frontend: Vue 3 + Vite + Pinia + Vue Router
- Database: MongoDB
- Cache/Lock: Redis
- Realtime: WebSocket (Gorilla)
- Message Queue: RabbitMQ
- Auth: Firebase Auth (Google Sign-In)

## Redis Distributed Lock

Prevents double booking when multiple users select the same seat at the same time.

```
Key:     lock:seat:{showtime_id}:{seat_no}
Acquire: SET key {token} NX EX 300
Release: Lua script checks token before DEL
```

Flow:

1. `POST /api/showtimes/:id/seats/lock` acquires Redis lock for each seat first
2. MongoDB seat status is updated to `LOCKED`
3. `POST /api/bookings/:id/pay` and cancel/timeout release Redis locks

## Phase 5 Test (Concurrency)

```powershell
$TOKEN = "your-jwt"
$SHOWTIME_ID = "your-showtime-id"
$BASE = "http://localhost:8080"

1..10 | ForEach-Object -Parallel {
  Invoke-WebRequest -Method POST "$using:BASE/api/showtimes/$using:SHOWTIME_ID/seats/lock" `
    -Headers @{ Authorization = "Bearer $using:TOKEN" } `
    -ContentType "application/json" `
    -Body '{"seat_nos":["A5"]}' `
    -UseBasicParsing
} -ThrottleLimit 10
```

Expected: exactly one `201 Created`, the rest `409 Conflict`.

## Phase 6 Test (WebSocket Real-time)

1. Restart backend and frontend
2. Open `http://localhost:5173/` → pick a showtime → **View seats**
3. Open the **same seat map URL** in a second browser tab (or incognito)
4. In tab A: select a seat → **Confirm seat selection**
5. Tab B should show that seat turn **yellow (LOCKED)** without refreshing

Ensure `frontend/.env` includes:

```
VITE_WS_URL=ws://localhost:8080
```

## RabbitMQ + Audit Logs (Phase 7)

Queues:

| Queue | Trigger | Audit event |
|-------|---------|-------------|
| `booking.success` | Pay succeeds | `BOOKING_SUCCESS` |
| `booking.timeout` | Lock expires | `BOOKING_TIMEOUT` |
| `seat.released` | Cancel booking | `SEAT_RELEASED` |

`SYSTEM_ERROR` is written directly to MongoDB when Redis lock infrastructure fails.

## Phase 7 Test

1. Set `RABBITMQ_URL` in `backend/.env` (CloudAMQP `amqps://...`)
2. Restart backend: `go run ./cmd/server`
3. Lock seats → Pay:
   ```powershell
   curl -X POST "$BASE/api/bookings/$BOOKING_ID/pay" -H "Authorization: Bearer $TOKEN"
   ```
4. Check server log for `mock notification: booking ... confirmed`
5. Open MongoDB Compass → `audit_logs` collection → see `BOOKING_SUCCESS`
6. CloudAMQP dashboard → Queues → messages published/consumed on `booking.success`
