# 🎬 Cinema Ticket Booking System — Flow & ขั้นตอนการทำ (0 → 100)

> เอกสารนี้คือ **แผนการทำงานแบบ Step-by-Step** สำหรับ Take-Home Assignment "ระบบจองตั๋วหนังออนไลน์"
> เปรียบเสมือนการสร้าง **"ตู้ขายตั๋วหนังออนไลน์"** ที่ **ไม่ใช่แค่รันได้** แต่ต้อง **ไม่พังเมื่อมีคนแย่งกันซื้อ**
>
> โฟกัสของโจทย์คือ **"ความคิด + ความถูกต้อง" มากกว่า Feature เยอะ ๆ**

---

## 📌 สรุปโจทย์ (อ่านก่อนเริ่ม)

### วัตถุประสงค์ของการประเมิน
- Backend / Frontend Engineering
- Concurrency & Distributed System
- System Design & DevOps
- คุณภาพโค้ดและการอธิบายแนวคิด

### ⏱ ระยะเวลา
**3 วัน (72 ชั่วโมง)** นับจากเวลาที่เริ่มงาน

### 🧱 Tech Stack ที่เราใช้ (ล็อกแล้ว — ไม่เปลี่ยน)

| Layer | Technology | Library / Package ที่ใช้จริง |
|---|---|---|
| Backend | **Go + Gin** | `github.com/gin-gonic/gin` |
| Frontend | **Vue 3** | Vue Router + Pinia + Vite |
| Database | **MongoDB** | `go.mongodb.org/mongo-driver` |
| Cache / Lock | **Redis** | `github.com/redis/go-redis/v9` (Distributed Lock เท่านั้น) |
| Realtime | **WebSocket** | `github.com/gorilla/websocket` + Gin |
| Message Queue | **RabbitMQ** | `github.com/rabbitmq/amqp091-go` |
| Auth | **Firebase Auth** | Frontend: `firebase` / Backend: `firebase.google.com/go/v4` |
| Deployment | **Docker + docker-compose** | รันทั้งระบบด้วย `docker compose up --build` |

### 🎯 Evaluation Criteria (น้ำหนักการให้คะแนน)

| หมวด | น้ำหนัก |
|---|---|
| Architecture & Concurrency Design | **35%** |
| Correctness (Booking / Lock / Realtime) | **30%** |
| DevOps & Security | **20%** |
| Code Quality & README | **15%** |

### ❌ สิ่งที่ "ไม่" คาดหวัง (อย่าเสียเวลากับสิ่งเหล่านี้)
- UI สวยระดับ Production
- Payment Gateway จริง (mock ได้)
- Feature ครบเหมือนระบบโรงหนังจริง

---

# 📚 คู่มือ Stack แต่ละตัว (อ่านก่อนเขียนโค้ด — สำคัญมาก)

> ถ้าไม่รู้จัก stack หลายตัว ให้อ่านส่วนนี้ก่อน แล้วค่อยไปทำ Phase
> เป้าหมายคือ **เข้าใจพอทำงานได้ใน 3 วัน** ไม่ต้องเชี่ยวชาญทุกอย่าง

---

## 1️⃣ Go + Gin (Backend Framework)

### Gin คืออะไร?
- **Gin** = Web Framework ของ Go สำหรับสร้าง REST API
- คล้าย Express.js (Node) หรือ Flask (Python) แต่เร็วกว่า
- ใช้จัดการ: route, middleware, request/response JSON

### ติดตั้ง & เริ่มต้น
```bash
# ติดตั้ง Go: https://go.dev/dl/
go version

mkdir cinema-booking/backend && cd cinema-booking/backend
go mod init github.com/yourname/cinema-booking

go get github.com/gin-gonic/gin
go get github.com/joho/godotenv   # โหลด .env
```

### โค้ดตัวอย่างขั้นต่ำ (ทดสอบว่า Gin รันได้)
```go
// cmd/server/main.go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    r.Run(":8080") // listen port 8080
}
```
```bash
go run cmd/server/main.go
# เปิด http://localhost:8080/health → {"status":"ok"}
```

### แนวคิดที่ต้องรู้
| แนวคิด | คืออะไร | ใช้ในโปรเจกต์นี้ยังไง |
|---|---|---|
| **Route** | URL + HTTP method | `GET /showtimes/:id/seats` |
| **Handler** | ฟังก์ชันรับ request | lock seat, pay booking |
| **Middleware** | ฟังก์ชันคั่นก่อน handler | ตรวจ JWT, ตรวจ role ADMIN |
| **Context (`c *gin.Context`)** | อ่าน param/body, ส่ง response | `c.Param("id")`, `c.JSON()` |
| **Group** | จัดกลุ่ม route | `/admin` ใส่ middleware ADMIN |

### Middleware ตัวอย่าง (Auth)
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization") // "Bearer xxx"
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        // verify JWT → set user_id ลง context
        c.Set("user_id", "xxx")
        c.Next() // ผ่านไป handler ถัดไป
    }
}

// ใช้งาน
api := r.Group("/api", AuthMiddleware())
api.POST("/bookings/:id/pay", payHandler)
```

### Package ที่ต้องติดตั้งทั้งหมด (Backend)
```bash
go get github.com/gin-gonic/gin
go get github.com/joho/godotenv
go get go.mongodb.org/mongo-driver/mongo
go get github.com/redis/go-redis/v9
go get github.com/gorilla/websocket
go get github.com/rabbitmq/amqp091-go
go get firebase.google.com/go/v4
go get github.com/golang-jwt/jwt/v5
```

---

## 2️⃣ MongoDB (Database)

### MongoDB คืออะไร?
- Database แบบ **NoSQL** เก็บข้อมูลเป็น **Document (JSON-like)**
- ไม่มี table แบบ SQL แต่มี **Collection** (เช่น `users`, `bookings`)
- เหมาะกับข้อมูลที่ structure ยืดหยุ่น

### แนวคิดพื้นฐาน
| SQL | MongoDB |
|---|---|
| Database | Database |
| Table | Collection |
| Row | Document |
| Column | Field |

### เชื่อมต่อจาก Go
```go
import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(uri string) (*mongo.Database, error) {
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
    if err != nil {
        return nil, err
    }
    return client.Database("cinema"), nil
}
```

### CRUD ตัวอย่าง
```go
// Insert
col := db.Collection("users")
_, err := col.InsertOne(ctx, bson.M{
    "email": "a@b.com",
    "role":  "USER",
})

// Find one
var user User
err := col.FindOne(ctx, bson.M{"email": "a@b.com"}).Decode(&user)

// Update
col.UpdateOne(ctx,
    bson.M{"_id": userID},
    bson.M{"$set": bson.M{"role": "ADMIN"}},
)

// Find many (filter bookings by movie)
cursor, err := col.Find(ctx, bson.M{"movie_id": movieID})
```

### เครื่องมือดูข้อมูล
- **MongoDB Compass** (GUI): https://www.mongodb.com/products/compass
- หรือเข้า mongo shell ใน Docker: `docker exec -it <mongo_container> mongosh`

### Connection string ใน Docker
```
mongodb://mongodb:27017/cinema
```
> `mongodb` = ชื่อ service ใน docker-compose (ไม่ใช่ localhost ตอนรันใน container)

---

## 3️⃣ Redis (Distributed Lock — หัวใจของโจทย์)

### Redis คืออะไร?
- Database แบบ **in-memory** เร็วมาก เก็บ key-value
- โจทย์บังคับ: ใช้ทำ **Distributed Lock** เท่านั้น (ไม่ใช่แค่ cache)

### ทำไมต้องใช้ Redis Lock?
```
สถานการณ์: User A และ User B กดจองที่นั่ง A5 พร้อมกัน

❌ ไม่มี Lock:
   A อ่าน MongoDB → AVAILABLE
   B อ่าน MongoDB → AVAILABLE  (พร้อมกัน!)
   A เขียน LOCKED
   B เขียน LOCKED  → Double Booking!

✅ มี Redis Lock:
   A → SET lock:seat:xxx:A5 tokenA NX EX 300 → สำเร็จ
   B → SET lock:seat:xxx:A5 tokenB NX EX 300 → ล้มเหลว (key มีอยู่แล้ว)
   B ได้ 409 "ที่นั่งถูกเลือกแล้ว"
```

### คำสั่ง Redis ที่ต้องรู้
| คำสั่ง | ความหมาย |
|---|---|
| `SET key value NX EX 300` | สร้าง key ได้เฉพาะถ้ายังไม่มี + หมดอายุ 300 วินาที (5 นาที) |
| `GET key` | อ่านค่า |
| `DEL key` | ลบ key (ปล่อย lock) |

### โค้ด Go (go-redis)
```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "redis:6379", // ชื่อ service ใน docker-compose
})

// Acquire lock
key := "lock:seat:showtime123:A5"
token := uuid.New().String()
ok, err := rdb.SetNX(ctx, key, token, 5*time.Minute).Result()
if !ok {
    // มีคน lock อยู่แล้ว → return 409
}

// Release lock อย่างปลอดภัย (Lua script)
script := `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end`
rdb.Eval(ctx, script, []string{key}, token)
```

### ทดสอบ Redis ด้วย CLI
```bash
docker exec -it <redis_container> redis-cli
SET test hello
GET test
SET lock:seat:1:A5 token123 NX EX 300
SET lock:seat:1:A5 token456 NX EX 300   # ตัวที่ 2 จะ fail
```

---

## 4️⃣ WebSocket (Real-time Seat Map)

### WebSocket คืออะไร?
- การเชื่อมต่อ **ถาวร** ระหว่าง browser กับ server (ต่างจาก HTTP ที่ request แล้วจบ)
- Server **push** ข้อมูลไป client ได้ทันที → ใช้ update ผังที่นั่ง real-time

### Flow ในโปรเจกต์นี้
```
1. User เข้าหน้า Seat Map → Frontend connect WebSocket
2. Join room ตาม showtime_id
3. User A lock ที่นั่ง A5 → Backend broadcast ไปทุกคนใน room
4. User B เห็น A5 เปลี่ยนเป็น LOCKED ทันที (ไม่ต้อง refresh)
```

### Backend (Gin + Gorilla WebSocket)
```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // dev only
}

func wsHandler(c *gin.Context) {
    conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
    defer conn.Close()
    showtimeID := c.Query("showtime_id")
    hub.Register(showtimeID, conn) // เข้า room
    for {
        _, _, err := conn.ReadMessage()
        if err != nil { break }
    }
    hub.Unregister(showtimeID, conn)
}

// Broadcast เมื่อ seat เปลี่ยน
hub.Broadcast(showtimeID, `{"type":"SEAT_UPDATE","seat_no":"A5","status":"LOCKED"}`)
```

### Frontend (Vue)
```js
const ws = new WebSocket(`ws://localhost:8080/ws?showtime_id=${id}`)
ws.onmessage = (event) => {
  const data = JSON.parse(event.data)
  if (data.type === 'SEAT_UPDATE') {
    seatStore.updateSeat(data.seat_no, data.status)
  }
}
```

---

## 5️⃣ RabbitMQ (Message Queue)

### RabbitMQ คืออะไร?
- **Message Queue** = กล่องข้อความระหว่าง service
- **Producer** ส่ง message → **Queue** → **Consumer** รับไปทำงาน
- ทำงาน **แบบ async** (ไม่ต้องรอให้เสร็จก่อนตอบ user)

### ทำไมต้องใช้ในโปรเจกต์นี้?
```
User จ่ายเงินสำเร็จ
  → API ตอบ 200 ทันที (เร็ว)
  → ส่ง message ไป RabbitMQ queue "booking.success"
  → Worker รับ message แล้ว:
      1. Mock ส่ง notification (log ว่า "ส่ง email แล้ว")
      2. เขียน audit_log ลง MongoDB
```

### แนวคิดพื้นฐาน
| คำศัพท์ | ความหมาย |
|---|---|
| **Exchange** | จุดรับ message จาก producer |
| **Queue** | คิวเก็บ message รอ consumer |
| **Binding** | เชื่อม exchange กับ queue |
| **Routing Key** | ชื่อ event เช่น `booking.success` |

### โค้ด Go (Producer)
```go
import amqp "github.com/rabbitmq/amqp091-go"

conn, _ := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
ch, _ := conn.Channel()
ch.QueueDeclare("booking.success", true, false, false, false, nil)

ch.PublishWithContext(ctx, "", "booking.success", false, false, amqp.Publishing{
    ContentType: "application/json",
    Body:        []byte(`{"booking_id":"xxx","user_id":"yyy"}`),
})
```

### โค้ด Go (Consumer — รันเป็น goroutine)
```go
msgs, _ := ch.Consume("booking.success", "", true, false, false, false, nil)
go func() {
    for msg := range msgs {
        // 1. log notification mock
        log.Println("📧 Mock email sent for booking:", string(msg.Body))
        // 2. เขียน audit_log ลง MongoDB
        auditRepo.Insert(ctx, "BOOKING_SUCCESS", string(msg.Body))
    }
}()
```

### RabbitMQ Management UI
- เปิดได้ที่ `http://localhost:15672` (user/pass: `guest`/`guest`)
- ใช้ดู queue, message ที่ค้างอยู่

---

## 6️⃣ Firebase Auth (Login ด้วย Google)

### Firebase Auth คืออะไร?
- Service ของ Google จัดการ login ให้ (Google Sign-In, email/password ฯลฯ)
- Frontend login → ได้ **ID Token** → ส่งให้ Backend verify → Backend ออก **JWT ของระบบเรา**

### Flow ที่ใช้ในโปรเจกต์
```
[Vue Frontend]                    [Go Backend]                  [MongoDB]
     │                                  │                            │
     │ 1. กด Login with Google          │                            │
     │ 2. Firebase popup → ID Token     │                            │
     │ 3. POST /auth/login {idToken} ──►│ 4. Verify token กับ Firebase│
     │                                  │ 5. Upsert user ───────────►│
     │                                  │ 6. สร้าง JWT ของเรา         │
     │◄── 7. คืน {access_token, user} ─│                            │
     │ 8. เก็บ token ใน Pinia/localStorage                           │
     │ 9. ทุก API แนบ Header: Authorization: Bearer <token>         │
```

### ขั้นตอนตั้งค่า Firebase (ทำครั้งเดียว)
1. ไป https://console.firebase.google.com → สร้าง Project
2. **Authentication** → Sign-in method → เปิด **Google**
3. **Project Settings** → General → เพิ่ม Web App → copy `firebaseConfig`
4. **Project Settings** → Service accounts → Generate new private key → ได้ไฟล์ JSON สำหรับ Backend

### Frontend (Vue)
```bash
npm install firebase
```
```js
// src/firebase.js
import { initializeApp } from 'firebase/app'
import { getAuth, GoogleAuthProvider, signInWithPopup } from 'firebase/auth'

const app = initializeApp({
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
})
export const auth = getAuth(app)

export async function loginWithGoogle() {
  const result = await signInWithPopup(auth, new GoogleAuthProvider())
  return result.user.getIdToken() // ส่ง token นี้ไป Backend
}
```

### Backend (Go) — Verify ID Token
```go
import firebase "firebase.google.com/go/v4"
import "firebase.google.com/go/v4/auth"

app, _ := firebase.NewApp(ctx, nil, opt.WithCredentialsFile("firebase-key.json"))
authClient, _ := app.Auth(ctx)
token, err := authClient.VerifyIDToken(ctx, idTokenFromFrontend)
uid := token.UID
email := token.Claims["email"]
```

### สร้าง JWT ของระบบเรา (หลัง verify สำเร็จ)
```go
import "github.com/golang-jwt/jwt/v5"

claims := jwt.MapClaims{
    "user_id": userID,
    "role":    "USER",
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signed, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
```

---

## 7️⃣ Vue 3 (Frontend)

### สร้างโปรเจกต์
```bash
npm create vue@latest
# เลือก: TypeScript (optional), Vue Router ✅, Pinia ✅
cd frontend && npm install && npm run dev
```

### โครงสร้างที่ใช้
```
src/
├── views/
│   ├── LoginView.vue
│   ├── ShowtimeListView.vue
│   ├── SeatMapView.vue
│   ├── PaymentView.vue
│   └── AdminView.vue
├── components/
│   └── SeatGrid.vue
├── stores/
│   ├── auth.js      # token, user, login/logout
│   ├── seatmap.js   # seats[], ws connection
│   └── booking.js   # current booking, countdown
├── api/
│   └── client.js    # axios + แนบ JWT
└── firebase.js
```

### API Client (แนบ JWT ทุก request)
```js
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL })

api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) config.headers.Authorization = `Bearer ${auth.token}`
  return config
})
export default api
```

---

## 8️⃣ Docker + docker-compose

### Docker คืออะไร?
- รันแอปใน **container** แยกจากเครื่องจริง
- ทุกคนรัน environment เดียวกัน → กรรมการรันได้ทันที

### docker-compose.yml ตัวอย่าง (ทั้งระบบ)
```yaml
services:
  mongodb:
    image: mongo:7
    ports: ["27017:27017"]
    volumes: [mongo_data:/data/db]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports: ["5672:5672", "15672:15672"]

  backend:
    build: ./backend
    ports: ["8080:8080"]
    env_file: ./backend/.env
    depends_on: [mongodb, redis, rabbitmq]

  frontend:
    build: ./frontend
    ports: ["3000:80"]
    depends_on: [backend]

volumes:
  mongo_data:
```

### คำสั่งที่ใช้บ่อย
```bash
docker compose up --build        # build + รันทุก service
docker compose up -d --build     # รัน background
docker compose down              # หยุดทุก service
docker compose logs backend -f   # ดู log backend
docker compose ps                # ดู service ที่รันอยู่
```

---

# 🗺️ ภาพรวมเส้นทาง 3 วัน (Roadmap)

```
วันที่ 1 (≈ 24 ชม.)
├── Phase 0   วางแผน + อ่านคู่มือ Stack       (4 ชม.)
├── Phase 1   Setup + Docker skeleton         (4 ชม.)
├── Phase 2   MongoDB + Seed data             (4 ชม.)
├── Phase 3   Firebase Auth (FE + BE)         (6 ชม.)
└── Phase 4   Seat Map API + Booking Core     (6 ชม.)

วันที่ 2 (≈ 24 ชม.)  ← วันสำคัญที่สุด
├── Phase 5   Redis Distributed Lock          (8 ชม.) ★
├── Phase 6   WebSocket Real-time             (6 ชม.)
├── Phase 7   RabbitMQ + Audit Log            (6 ชม.)
└── Phase 8   Admin + RBAC                    (4 ชม.)

วันที่ 3 (≈ 24 ชม.)
├── Phase 9   Frontend Vue เชื่อมทุกอย่าง     (10 ชม.)
├── Phase 10  Docker Compose polish           (4 ชม.)
├── Phase 11  Concurrency Test                (4 ชม.)
└── Phase 12  README + ส่งงาน                 (6 ชม.)
```

> เคล็ดลับ: **วันที่ 2 ต้องจบ Phase 5 (Redis Lock)** ให้ได้ — นี่คือ 35% ของคะแนน

---

# 🏗️ System Architecture

```
                    ┌──────────────────────────────┐
                    │   Frontend (Vue 3 + Vite)     │
                    │   - Firebase Google Login     │
                    │   - Seat Map (WebSocket)      │
                    │   - Admin Dashboard           │
                    └──────────────┬───────────────┘
                                   │ REST API + WebSocket
                                   ▼
                    ┌──────────────────────────────┐
                    │   Backend (Go + Gin)          │
                    │   - Firebase Token Verify   │
                    │   - JWT Middleware + RBAC     │
                    │   - Booking Service         │
                    │   - WebSocket Hub (Gorilla)   │
                    │   - RabbitMQ Producer         │
                    │   - Redis Lock Manager        │
                    └──┬─────────┬─────────┬───────┘
                       │         │         │
           ┌───────────┘         │         └──────────────┐
           ▼                     ▼                        ▼
   ┌───────────────┐    ┌────────────────┐    ┌──────────────────┐
   │   MongoDB     │    │     Redis      │    │    RabbitMQ      │
   │ users         │    │ SET NX EX 300  │    │ queue:           │
   │ showtimes     │    │ (Distributed   │    │ booking.success  │
   │ seats         │    │  Lock 5 นาที)  │    │ booking.timeout  │
   │ bookings      │    └────────────────┘    └────────┬─────────┘
   │ audit_logs    │                                   │
   └───────────────┘                                   ▼
                                            ┌──────────────────┐
                                            │ Worker (goroutine)│
                                            │ - Mock notification│
                                            │ - Write audit_log │
                                            └──────────────────┘
```

---

# ✅ PHASE 0 — วางแผน & System Design (0% → 5%)

**เป้าหมาย:** เข้าใจโจทย์ + อ่านคู่มือ Stack ด้านบนให้จบ

### ขั้นตอน
1. อ่านโจทย์ให้จบ + ทำ Requirement Checklist
2. **อ่านคู่มือ Stack ทั้ง 8 ตัว** ด้านบน (โดยเฉพาะ Redis Lock + RabbitMQ)
3. ติดตั้งเครื่องมือบนเครื่อง:
   - Go 1.22+: https://go.dev/dl/
   - Node.js 20+: https://nodejs.org/
   - Docker Desktop: https://www.docker.com/products/docker-desktop/
   - MongoDB Compass (optional): ดูข้อมูลใน DB
4. สร้าง Firebase Project + เปิด Google Sign-In (ดูคู่มือ Firebase ด้านบน)
5. วาด State Machine ที่นั่ง:
   ```
   AVAILABLE ──(Lock สำเร็จ)──► LOCKED ──(จ่ายเงิน)──► BOOKED
       ▲                            │
       └──(TTL 5 นาที / ยกเลิก)────┘
   ```
6. กำหนด Environment Variables ทั้งหมด (ห้าม hardcode):

```env
# backend/.env
PORT=8080
MONGO_URI=mongodb://mongodb:27017/cinema
REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
JWT_SECRET=your-secret-key
FIREBASE_CREDENTIALS=/app/firebase-key.json
LOCK_TTL_SECONDS=300

# frontend/.env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
VITE_FIREBASE_API_KEY=xxx
VITE_FIREBASE_AUTH_DOMAIN=xxx.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=xxx
```

**Output:** เข้าใจ stack + มี Firebase project + รายการ env ครบ

---

# ✅ PHASE 1 — Setup โครงสร้าง + Docker Skeleton (5% → 15%)

**เป้าหมาย:** `docker compose up --build` แล้วทุก container ขึ้นได้

### โครงสร้างโปรเจกต์
```
cinema-booking/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/          # โหลด env
│   │   ├── auth/            # Firebase verify + JWT middleware
│   │   ├── booking/         # booking service
│   │   ├── seat/            # seat map handler
│   │   ├── lock/            # Redis distributed lock
│   │   ├── realtime/        # WebSocket hub (Gorilla)
│   │   ├── mq/              # RabbitMQ producer + consumer
│   │   ├── admin/           # admin API + RBAC
│   │   └── repository/      # MongoDB repositories
│   ├── go.mod
│   ├── Dockerfile
│   ├── .env.example
│   └── firebase-key.json    # (gitignore — ใส่ใน Docker volume)
├── frontend/
│   ├── src/ ...
│   ├── Dockerfile
│   └── .env.example
├── docker-compose.yml
└── README.md
```

### ขั้นตอน
1. `git init` + `.gitignore` (node_modules, .env, firebase-key.json)
2. สร้าง Go module + ติดตั้ง package ทั้งหมด (ดูรายการในคู่มือ Gin)
3. สร้าง Vue 3 project: `npm create vue@latest` (เลือก Router + Pinia)
4. เขียน `docker-compose.yml` (copy จากคู่มือ Docker ด้านบน)
5. เขียน `backend/Dockerfile`:
   ```dockerfile
   FROM golang:1.22-alpine AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN CGO_ENABLED=0 go build -o server ./cmd/server

   FROM alpine:latest
   COPY --from=builder /app/server /server
   COPY firebase-key.json /app/firebase-key.json
   CMD ["/server"]
   ```
6. เขียน `frontend/Dockerfile` (build Vue → serve ด้วย nginx)
7. สร้าง `GET /health` ใน Gin → ทดสอบ `docker compose up --build`

**Output:** ทุก service ขึ้น + `/health` ตอบ 200

---

# ✅ PHASE 2 — MongoDB + Seed Data (15% → 25%)

**เป้าหมาย:** มีข้อมูลหนัง/รอบฉาย/ที่นั่งให้เล่นได้

### Collection Schema
```js
// users
{ _id, firebase_uid, email, name, role: "USER"|"ADMIN", created_at }

// movies
{ _id, title, poster_url, duration_min }

// showtimes
{ _id, movie_id, start_time, screen, price }

// seats (1 doc ต่อ 1 ที่นั่งต่อ 1 รอบ)
{ _id, showtime_id, seat_no, status: "AVAILABLE"|"LOCKED"|"BOOKED",
  locked_by, locked_at, lock_token, booking_id }

// bookings
{ _id, user_id, showtime_id, seat_nos: [], status: "PENDING"|"PAID"|"EXPIRED",
  amount, created_at, paid_at, expires_at }

// audit_logs
{ _id, event: "BOOKING_SUCCESS"|"BOOKING_TIMEOUT"|"SEAT_RELEASED"|"SYSTEM_ERROR",
  user_id, showtime_id, seat_no, detail, created_at }
```

### ขั้นตอน
1. เขียน `internal/repository/` แยกตาม collection (users, seats, bookings, audit)
2. เขียน `internal/config/` โหลด `MONGO_URI` จาก env
3. สร้าง index:
   ```go
   seatsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
       Keys: bson.D{{"showtime_id", 1}, {"seat_no", 1}},
       Options: options.Index().SetUnique(true),
   })
   ```
4. เขียน seed script (`cmd/seed/main.go`):
   - สร้าง 1 หนัง, 2 รอบฉาย
   - สร้างที่นั่ง A1–J10 (10 แถว × 10 ที่) ต่อรอบ = 100 ที่นั่ง
   - สร้าง admin user 1 คน (`role: "ADMIN"`)
5. รัน seed: `go run cmd/seed/main.go` (หรือรันอัตโนมัติตอน backend start)
6. เปิด MongoDB Compass ดูข้อมูล

**Output:** มีข้อมูลจริงใน MongoDB

---

# ✅ PHASE 3 — Firebase Auth (25% → 35%)

**เป้าหมาย:** Login ด้วย Google ได้ + มี JWT ของระบบเรา

### API
```
POST /api/auth/login    Body: { "id_token": "..." }  → { access_token, user }
GET  /api/me            Header: Bearer <token>        → { user_id, email, role }
```

### ขั้นตอน (ทำตามลำดับ)
1. **Firebase Console:** เปิด Google Sign-In (ดูคู่มือ Firebase)
2. **Frontend:** สร้าง `src/firebase.js` + ปุ่ม Login ใน `LoginView.vue`
3. **Frontend:** หลัง login สำเร็จ → ส่ง `id_token` ไป `POST /api/auth/login`
4. **Backend:** สร้าง `internal/auth/firebase.go` — verify ID token ด้วย Firebase Admin SDK
5. **Backend:** upsert user ลง MongoDB (`firebase_uid`, `email`, `role: USER`)
6. **Backend:** สร้าง JWT ของเรา (`user_id`, `role`, `exp 24h`)
7. **Backend:** เขียน `AuthMiddleware()` ตรวจ JWT ทุก protected route
8. **Backend:** เขียน `RequireRole("ADMIN")` middleware
9. **ทดสอบ:** login → ได้ token → `GET /api/me` คืนข้อมูล user

**Output:** ระบบรู้ว่าใคร login + มี role

---

# ✅ PHASE 4 — Seat Map API + Booking Core (35% → 50%)

**เป้าหมาย:** API ดูผังที่นั่ง + เริ่ม flow จอง (ยังไม่มี Redis lock)

### API
```
GET  /api/showtimes                    # รายการรอบฉาย
GET  /api/showtimes/:id/seats          # ผังที่นั่งทั้งหมด
POST /api/showtimes/:id/seats/lock     # เลือกที่นั่ง (จะผูก Redis ใน Phase 5)
POST /api/bookings/:id/pay             # จ่ายเงิน mock → BOOKED
POST /api/bookings/:id/cancel          # ยกเลิก → ปล่อย lock
GET  /api/bookings/my                  # ประวัติการจองของ user
```

### Booking Flow
```
1. User เลือกที่นั่ง → POST /seats/lock { seat_nos: ["A5"] }
2. Backend สร้าง booking status=PENDING, expires_at = now + 5 นาที
3. User กดจ่ายเงิน → POST /bookings/:id/pay
4. Backend เปลี่ยน seat=BOOKED, booking=PAID
5. ถ้าไม่จ่ายใน 5 นาที → booking=EXPIRED, seat=AVAILABLE
```

### ขั้นตอน
1. เขียน `GET /showtimes` + `GET /showtimes/:id/seats`
2. เขียน `POST /seats/lock` (update MongoDB อย่างเดียวก่อน — ยังไม่กัน race)
3. เขียน `POST /bookings/:id/pay` (mock: ไม่ต้อง payment gateway จริง)
4. เขียน `POST /bookings/:id/cancel`
5. เขียน background goroutine ตรวจ booking `PENDING` ที่ `expires_at` ผ่านแล้ว → set EXPIRED
6. ทดสอบด้วย Postman/curl ทีละ step

**Output:** จองได้ครบ flow (ยังไม่กัน concurrent)

---

# ✅ PHASE 5 — Redis Distributed Lock (50% → 65%) ★ สำคัญที่สุด

**เป้าหมาย:** กัน Double Booking เมื่อหลายคนกดพร้อมกัน

### Lock Strategy (ต้องเขียนใน README)
```
Key:    lock:seat:{showtime_id}:{seat_no}
Command: SET key {token} NX EX 300
Release: Lua script เช็ค token ก่อน DEL
```

### ขั้นตอน
1. สร้าง `internal/lock/redis_lock.go`:
   - `AcquireLock(ctx, showtimeID, seatNo, userID) (token string, err error)`
   - `ReleaseLock(ctx, showtimeID, seatNo, token) error`
2. แก้ `POST /seats/lock`:
   ```
   for each seat_no:
     token, err := lock.AcquireLock(...)
     if err → return 409 "ที่นั่ง X ถูกเลือกแล้ว"
   update MongoDB seats → LOCKED
   create booking PENDING
   ```
3. แก้ `POST /bookings/:id/pay`:
   ```
   verify booking ยัง PENDING + ยังไม่หมดเวลา
   update seats → BOOKED
   ReleaseLock ทุก seat
   update booking → PAID
   ```
4. แก้ `POST /bookings/:id/cancel` + timeout handler → ReleaseLock
5. Background worker: ตรวจ booking PENDING หมดเวลา → ReleaseLock + log `BOOKING_TIMEOUT`
6. **ทดสอบ concurrency** (Phase 11 จะทำอีกรอบ แต่ลองเบื้องต้นที่นี่):
   ```bash
   # ยิง 10 requests พร้อมกัน
   for i in {1..10}; do
     curl -X POST http://localhost:8080/api/showtimes/xxx/seats/lock \
       -H "Authorization: Bearer $TOKEN" \
       -d '{"seat_nos":["A5"]}' &
   done
   wait
   # ต้องได้สำเร็จแค่ 1, ที่เหลือ 409
   ```

**Output:** ไม่มี Double Booking

---

# ✅ PHASE 6 — WebSocket Real-time (65% → 75%)

**เป้าหมาย:** ทุกคนเห็น seat เปลี่ยนสถานะทันที

### ขั้นตอน
1. สร้าง `internal/realtime/hub.go`:
   - `Register(showtimeID, conn)` / `Unregister(showtimeID, conn)`
   - `Broadcast(showtimeID, message)` — ส่งไปทุก conn ใน room
2. เพิ่ม route `GET /ws?showtime_id=xxx` upgrade เป็น WebSocket (Gorilla)
3. เรียก `hub.Broadcast()` ทุกครั้งที่ seat status เปลี่ยน (lock / pay / cancel / timeout):
   ```json
   { "type": "SEAT_UPDATE", "seat_no": "A5", "status": "LOCKED" }
   ```
4. **Frontend:** ใน `SeatMapView.vue` connect WebSocket ตอน mount:
   ```js
   const ws = new WebSocket(`${import.meta.env.VITE_WS_URL}/ws?showtime_id=${id}`)
   ws.onmessage = (e) => seatStore.handleUpdate(JSON.parse(e.data))
   ```
5. ทดสอบ: เปิด 2 browser tab → tab A lock ที่นั่ง → tab B เห็นเปลี่ยนทันที

**Output:** Seat map real-time ทำงาน

---

# ✅ PHASE 7 — RabbitMQ + Audit Logs (75% → 82%)

**เป้าหมาย:** MQ ถูกใช้งานจริง + audit log ครบ 4 event

### Queue ที่ใช้
| Queue | เมื่อไหร่ส่ง | Consumer ทำอะไร |
|---|---|---|
| `booking.success` | จ่ายเงินสำเร็จ | mock notification + log `BOOKING_SUCCESS` |
| `booking.timeout` | lock หมดเวลา | log `BOOKING_TIMEOUT` |
| `seat.released` | ยกเลิกจอง | log `SEAT_RELEASED` |

### ขั้นตอน
1. สร้าง `internal/mq/rabbitmq.go` — connect, declare queues
2. สร้าง `internal/mq/producer.go` — `PublishBookingSuccess(booking)`
3. สร้าง `internal/mq/consumer.go` — รัน goroutine ตอน server start:
   ```go
   go consumer.Start(ctx) // listen booking.success → mock email + audit log
   ```
4. เรียก producer ใน `pay` handler หลัง PAID สำเร็จ
5. เรียก producer ใน timeout/cancel handler
6. เขียน `internal/admin/audit.go` — insert audit_log ลง MongoDB
7. Log `SYSTEM_ERROR` เมื่อ Redis lock fail
8. เปิด RabbitMQ UI (`http://localhost:15672`) ดู message ไหล

**Output:** MQ ทำงานจริง + audit log ครบ

---

# ✅ PHASE 8 — Admin Dashboard + RBAC (82% → 88%)

**เป้าหมาย:** Admin ดู booking + filter ได้, User เข้าไม่ได้

### API (ต้องมี middleware `RequireRole("ADMIN")`)
```
GET /api/admin/bookings?movie_id=&date=&user_id=   # filter อย่างน้อย 1
GET /api/admin/audit-logs?event=&from=&to=
```

### ขั้นตอน
1. ใส่ `RequireRole("ADMIN")` ทุก route `/api/admin/*`
2. Query bookings + join movie name สำหรับ filter
3. รองรับ filter: `?movie_id=xxx` หรือ `?date=2026-06-11` หรือ `?user_id=xxx`
4. ทดสอบ: USER token → `GET /api/admin/bookings` → **403 Forbidden**
5. ADMIN token → ได้รายการ

**Output:** Admin API ปลอดภัย

---

# ✅ PHASE 9 — Frontend Vue 3 เชื่อมทุกอย่าง (88% → 95%)

**เป้าหมาย:** UI ใช้งานได้ครบ flow

### หน้าที่ต้องมี
| หน้า | Route | ทำอะไร |
|---|---|---|
| Login | `/login` | ปุ่ม Google login (Firebase) |
| Showtimes | `/` | รายการรอบฉาย |
| Seat Map | `/showtimes/:id` | ผังที่นั่ง + WebSocket real-time |
| Payment | `/bookings/:id/pay` | countdown 5 นาที + ปุ่มจ่าย mock |
| My Bookings | `/bookings` | ประวัติจอง |
| Admin | `/admin` | ตาราง booking + filter + audit logs |

### สีที่นั่ง (ใช้ CSS class ไม่ใช้ inline style)
```css
.seat-available { background: #4ade80; }
.seat-locked    { background: #facc15; }
.seat-booked    { background: #f87171; }
```

### ขั้นตอน
1. สร้าง Pinia stores: `auth`, `seatmap`, `booking`
2. สร้าง `api/client.js` แนบ JWT (ดูคู่มือ Vue)
3. ทำ Login flow: Firebase popup → ส่ง id_token → เก็บ access_token
4. ทำ SeatMap: โหลด seats จาก API + connect WebSocket
5. ทำ Payment: countdown 5 นาที (setInterval) + ปุ่ม "จ่ายเงิน (Mock)"
6. ทำ Admin: ตาราง + dropdown filter
7. Route guard: `/admin` เข้าได้เฉพาะ `role === 'ADMIN'`
8. ใช้ semantic HTML (`<section>`, `<article>`, `<header>`), alt ทุก `<img>`

**Output:** จองตั๋วผ่าน UI ได้จริง end-to-end

---

# ✅ PHASE 10 — Docker Compose Polish (95% → 97%)

**เป้าหมาย:** `docker compose up --build` รันทั้งระบบได้คำสั่งเดียว

### ขั้นตอน
1. ใส่ **healthcheck** ทุก service:
   ```yaml
   mongodb:
     healthcheck:
       test: echo 'db.runCommand("ping").ok' | mongosh localhost:27017/test --quiet
   redis:
     healthcheck:
       test: ["CMD", "redis-cli", "ping"]
   rabbitmq:
     healthcheck:
       test: rabbitmq-diagnostics -q ping
   backend:
     depends_on:
       mongodb: { condition: service_healthy }
       redis: { condition: service_healthy }
       rabbitmq: { condition: service_healthy }
   ```
2. Backend รัน seed data อัตโนมัติตอน start (ถ้ายังไม่มีข้อมูล)
3. Frontend build ด้วย nginx serve static files
4. ทดสอบจากเครื่องเปล่า: clone → `docker compose up --build` → เปิด `http://localhost:3000`

**Output:** กรรมการรันได้ทันที

---

# ✅ PHASE 11 — Concurrency Test (97% → 99%)

**เป้าหมาย:** พิสูจน์ว่าไม่มี Double Booking

### ขั้นตอน
1. เขียน Go test (`internal/booking/lock_test.go`):
   ```go
   func TestConcurrentLock(t *testing.T) {
       var wg sync.WaitGroup
       success := 0
       for i := 0; i < 50; i++ {
           wg.Add(1)
           go func() {
               defer wg.Done()
               token, err := lock.AcquireLock(ctx, showtimeID, "A5", uuid.New().String())
               if err == nil { success++ }
           }()
       }
       wg.Wait()
       if success != 1 {
           t.Fatalf("expected 1 success, got %d", success)
       }
   }
   ```
2. ทดสอบ timeout: ตั้ง `LOCK_TTL_SECONDS=10` ชั่วคราว → lock แล้วไม่จ่าย → 10 วิ seat กลับ AVAILABLE
3. ทดสอบ WebSocket: 2 tab lock/pay → ทั้ง 2 tab เห็น update
4. (Optional) สร้าง Postman Collection

**Output:** มีหลักฐานว่าระบบถูกต้อง

---

# ✅ PHASE 12 — README + ส่งงาน (99% → 100%)

### README.md ต้องมี (ตามโจทย์)
1. **System Architecture Diagram**
2. **Tech Stack Overview** (ตารางด้านบน)
3. **Booking Flow** step-by-step
4. **Redis Lock Strategy** — อธิบาย `SET NX EX` + token + Lua + trade-off
5. **RabbitMQ Usage** — queue ไหน ส่งเมื่อไหร่ consumer ทำอะไร
6. **How to Run** — `docker compose up --build`
7. **Assumptions & Trade-offs**

### ขั้นสุดท้าย
1. เขียน README ครบ 7 หัวข้อ
2. ใส่ `.env.example` ทั้ง backend + frontend
3. ทดสอบ clone ใหม่ → `docker compose up --build` → ใช้งานได้
4. Push Git repository

**Output:** งานพร้อมส่ง 🎉

---

# 🧩 Checklist สรุป (เช็คก่อนส่ง)

- [ ] `docker compose up --build` รันทั้งระบบได้คำสั่งเดียว
- [ ] Login ด้วย Firebase Google ได้ + มี `user_id` ผูกการจอง
- [ ] Seat map แสดง AVAILABLE / LOCKED / BOOKED
- [ ] Seat map อัปเดต real-time ผ่าน WebSocket (Gorilla)
- [ ] Lock ด้วย Redis `SET NX EX 300` + ปล่อยด้วย Lua script
- [ ] ไม่จ่ายใน 5 นาที → seat กลับ AVAILABLE อัตโนมัติ
- [ ] จ่ายสำเร็จ → BOOKED
- [ ] ไม่มี Double Booking (มี test หลักฐาน)
- [ ] RabbitMQ ถูกใช้จริง (booking.success → mock notification + audit log)
- [ ] Audit log ครบ: BOOKING_SUCCESS, BOOKING_TIMEOUT, SEAT_RELEASED, SYSTEM_ERROR
- [ ] Admin Dashboard + filter อย่างน้อย 1 + RBAC (USER → 403)
- [ ] ไม่ hardcode config (ใช้ env ทั้งหมด)
- [ ] README ครบ 7 หัวข้อ + Architecture Diagram

---

# 🆘 Troubleshooting ที่เจอบ่อย

| ปัญหา | สาเหตุ | แก้ |
|---|---|---|
| Backend connect MongoDB ไม่ได้ | ใช้ `localhost` ใน Docker | ใช้ชื่อ service `mongodb:27017` |
| Redis lock ไม่ทำงาน | Redis ยังไม่ขึ้น | รอ healthcheck ผ่านก่อน |
| Firebase verify fail | ไฟล์ credentials ผิด path | mount `firebase-key.json` ใน Dockerfile |
| WebSocket disconnect บ่อย | ไม่มี reconnect logic | เพิ่ม `ws.onclose` → reconnect 3 วิ |
| RabbitMQ connection refused | service ยังไม่พร้อม | ใส่ `depends_on` + retry connect |
| CORS error | Frontend คนละ port | ใส่ CORS middleware ใน Gin |
| JWT invalid | secret ไม่ตรง / token หมดอายุ | ตรวจ `JWT_SECRET` + `exp` claim |

---

> 💡 **จำไว้:** Architecture & Concurrency (35%) + Correctness (30%) = **65%**
> ทุ่มเวลากับ **Redis Lock (Phase 5)** มากที่สุด — ถ้าส่วนนี้ถูกต้อง งานเกือบผ่านแล้ว
