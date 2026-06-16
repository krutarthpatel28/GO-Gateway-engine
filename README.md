# GO-Gateway-Engine 🚀

A high-performance, concurrent API Gateway proxy featuring a thread-safe in-memory fast cache and a token-bucket rate limiter. Built from scratch using native Go concurrency primitives, this engine achieves blazing-fast sub-50 microsecond execution times.

## 🛠️ System Architecture

The gateway acts as an intercepting proxy shield between incoming client requests and backend computation layers. It is composed of three decoupled, high-performance packages:

1. **The Fast-Cache Data Store (`/cache`):** A multi-tenant, nested hash-map layout (`map[string]map[string]string`) that isolates cached resources by `X-API-Key`. Fully optimized for concurrent reading using a Read-Write Mutex (`sync.RWMutex`).
2. **The Token-Bucket Rate Limiter (`/ratelimiter`):** An automated traffic gatekeeper tracking client request balances. Driven by a dedicated background worker thread (**Goroutine**) utilizing an asynchronous channel stream (`time.Ticker`) to replenish tokens every 10 seconds.
3. **The HTTP Execution Pipeline (`main.go`):** A modular routing server managing HTTP stream decoding, structured JSON validation, and non-blocking middleware injection.

---

## ⚡ Performance Metrics

Executing locally over loopback interfaces, the application layer processing times consistently clock in at microseconds:

* **Cache Write Ops (`POST /cache`):** ~30μs - 50μs
* **Cache Read Ops (`GET /cache`):** ~15μs - 25μs

---

## 🚦 Endpoint Specifications

All core routes are fully protected behind the `RateLimiterMiddleware`. Incoming requests must supply a valid identifier header or face immediate termination.

### 1. Cache Storage Pipeline

* **Route:** `POST /cache`
* **Headers:** `X-API-Key: <your_secret_string>`
* **Payload (JSON):**

```json
{
  "key": "session_id",
  "value": "premium_user_token_99"
}
```

* **Response Status:** `201 Created`

### 2. Cache Extraction Pipeline

* **Route:** `GET /cache?key=<target_key>`
* **Headers:** `X-API-Key: <your_secret_string>`
* **Response Status:** `200 OK`
* **Response Body:**

```json
{
  "key": "session_id",
  "value": "premium_user_token_99"
}
```

### 3. Rate Limiter Interception

* **Trigger:** Exceeding the maximum token capacity within a 10-second window.
* **Response Status:** `429 Too Many Requests`
* **Response Body:** `Too Many Requests`

---

## 🚀 Getting Started

### Prerequisites

* Go 1.22+ installed natively on your machine.

### Installation & Execution

Clone the repository into your local workspace:

```bash
git clone https://github.com/your-username/GO-Gateway-Engine.git
cd GO-Gateway-Engine
```

Spin up the server engine:

```bash
go run main.go
```

Verify the engine boots up successfully in your terminal:

```
🚀 Starting Gateway Engine on port 8080...
```

---

## 🧪 Verification via cURL

Open a separate terminal window and fire these test calls to see the memory maps and worker tickers coordinate in real time:

```bash
# 1. Populate the multi-tenant cache
curl -X POST http://localhost:8080/cache \
     -H "X-API-Key: user_krutarth" \
     -H "Content-Type: application/json" \
     -d '{"key": "theme", "value": "dark_mode"}' -i

# 2. Query your private data namespace
curl -X GET "http://localhost:8080/cache?key=theme" \
     -H "X-API-Key: user_krutarth" -i

# 3. Verify security isolation (different API key gets a 404)
curl -X GET "http://localhost:8080/cache?key=theme" \
     -H "X-API-Key: anonymous_attacker" -i

# 4. Stress test the rate limiter threshold (Trigger 429)
for i in {1..12}; do curl -X GET "http://localhost:8080/cache?key=theme" -H "X-API-Key: user_krutarth" -o /dev/null -s -w "Request $i: %{http_code}\n"; done
```

---

## 📝 Key Engineering Takeaways

* **Data-Race Prevention:** Solved Go's native map concurrent-write panic limitation by building isolation boundaries around memory critical sections using `sync.Mutex` and `sync.RWMutex`.
* **Resource Preservation:** Bypassed standard application loops (O(N) scanning costs) by implementing nested maps to locate multi-user configurations in flat constant time (O(1)).
* **Asynchronous Concurrency:** Leveraged Go's highly parallel scheduling model to offload background bucket calculations away from the main user-facing HTTP request thread.
