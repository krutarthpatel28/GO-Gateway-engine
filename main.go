package main

import (
	"encoding/json"
	"fmt"
	"gateway-engine/cache"
	"gateway-engine/rate-limiter"
	"net/http"
	"time"
)

type App struct{
	Port string
	RL *ratelimiter.RateKeeper
	Cache *cache.MemoryCache
}

func NewApp(port string, rl *ratelimiter.RateKeeper, cache *cache.MemoryCache) *App{
	return &App{
		Port: port,
		RL: rl,
		Cache: cache,
	}
}

type Records struct{
	Key string `json:"key"`
	Value string `json:"value"`
}

func RateLimiterMiddleware(next http.Handler, app *App) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		apiKey := r.Header.Get("X-API-Key")
		if apiKey=="" {
			http.Error(w,"API KEY MISSING", 401)
			return 
		}
		if !app.RL.RequestChecker(apiKey){
			http.Error(w,"Too Many Request", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		start := time.Now()
		next.ServeHTTP(w,r)
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s took %v\n", r.Method, r.URL, r.RemoteAddr, duration)
	})
}

// Routes 
func (app *App) CreateReqHandler(w http.ResponseWriter, r *http.Request){
	apiKey := r.Header.Get("X-API-Key")
	w.Header().Set("Content-Type","application/json")
	var rec Records
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
		return
	}
	app.Cache.Set(apiKey, rec.Key, rec.Value)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rec)
}

func (app *App) GetReqHandler(w http.ResponseWriter, r *http.Request){
	apiKey := r.Header.Get("X-API-Key")
	w.Header().Set("Content-Type","application/json")
	if apiKey=="" {
		http.Error(w,"API KEY MISSING", http.StatusUnauthorized)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing key parameter"}`))
		return
	}
	
	value, err := app.Cache.Get(apiKey, key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"status": "Failed", "message": err.Error()})
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Records{Key: key, Value: value})
}

// Routes with Middleware
func ProtectedRoute(next http.Handler, app *App) http.Handler{
	return RateLimiterMiddleware(LoggingMiddleware(next), app)
}

func main(){
	mux := http.NewServeMux()
	
	app := NewApp("8080", ratelimiter.NewRateLimiter(10), cache.NewMemoryCache())
	fmt.Printf("🚀 Starting Gateway Engine on port %s...\n", app.Port)

	mux.Handle("POST /cache",ProtectedRoute(http.HandlerFunc(app.CreateReqHandler), app))
	mux.Handle("GET /cache",ProtectedRoute(http.HandlerFunc(app.GetReqHandler), app))

	if err := http.ListenAndServe(":"+app.Port, mux); err != nil{
		fmt.Println("Error Starting server")
	}
}