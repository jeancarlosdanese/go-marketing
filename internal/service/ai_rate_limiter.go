// File: /internal/service/ai_rate_limiter.go

package service

import (
	"sync"
	"time"
)

// OpenAIRateLimiter controla as requisições para a OpenAI.
type OpenAIRateLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	lastReq  time.Time
}

// NewOpenAIRateLimiter cria um novo rate limiter baseado na taxa máxima de requisições permitida.
func NewOpenAIRateLimiter(rps int) *OpenAIRateLimiter {
	return &OpenAIRateLimiter{
		interval: time.Second / time.Duration(rps),
		lastReq:  time.Now(),
	}
}

// Wait aguarda até que seja seguro fazer a próxima requisição.
func (rl *OpenAIRateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastReq)

	if elapsed < rl.interval {
		time.Sleep(rl.interval - elapsed)
	}

	rl.lastReq = time.Now()
}
