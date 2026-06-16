package ratelimiter

import (
	"sync"
	"time"
	"fmt"
)

type RateKeeper struct{
	mu sync.Mutex
	Rate map[string]int
	maxToken int
}

func NewRateLimiter(max int) *RateKeeper{
	rl := &RateKeeper{
		Rate: make(map[string]int),
		maxToken: max,
	}
	go rl.startRefillTicker()
	return rl
}

func (limiter *RateKeeper) RequestChecker(apiKey string) bool{
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if _, exists := limiter.Rate[apiKey]; !exists {
		limiter.Rate[apiKey] = limiter.maxToken - 1
		return true
	}
	usrRate := limiter.Rate[apiKey]
	if usrRate > 0 {
		limiter.Rate[apiKey] = limiter.Rate[apiKey] - 1
		return true
	}else{
		return false
	}
}

func (limiter *RateKeeper) startRefillTicker(){
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C{
		limiter.mu.Lock()
		fmt.Println("🔄 RateLimiter Background Worker: Refilling all token buckets...")
		for key := range limiter.Rate{
			limiter.Rate[key] = limiter.maxToken
		}
		limiter.mu.Unlock()
	}
}