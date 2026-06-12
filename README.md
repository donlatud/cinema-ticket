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
