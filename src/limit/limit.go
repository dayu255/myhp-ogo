package limit

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type limiter struct {
	mu         sync.RWMutex
	limiter    *rate.Limiter
	lastUpdate time.Time
}

type IPLimitStore struct {
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
	maxIP    int
	limiters map[string]*limiter
	reducing atomic.Bool
}

func NewLimitStore(limit rate.Limit, burst int, maxIP int) *IPLimitStore {
	return &IPLimitStore{
		mu:       sync.RWMutex{},
		limit:    limit,
		burst:    burst,
		maxIP:    maxIP,
		limiters: make(map[string]*limiter),
	}
	// reducing atomic.Bool の初期値はfalse
}

func (l *IPLimitStore) Allow(ip string) (ok bool) {
	l.mu.RLock()
	lmt, exist := l.limiters[ip]
	IPcount := len(l.limiters)
	l.mu.RUnlock()

	if IPcount >= l.maxIP && l.reducing.CompareAndSwap(false, true) {
		go l.reduce()
	}

	if !exist {
		l.mu.Lock()
		lmt, exist = l.limiters[ip]
		if !exist {
			lmt = &limiter{
				limiter:    rate.NewLimiter(l.limit, l.burst),
				lastUpdate: time.Now(),
			}
			l.limiters[ip] = lmt
		}
		l.mu.Unlock()
	}

	lmt.mu.Lock()
	defer lmt.mu.Unlock()
	lmt.lastUpdate = time.Now()
	return lmt.limiter.Allow()
}

func (l *IPLimitStore) StartCleanup(ctx context.Context, interval time.Duration, deadtime time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.cleanup(deadtime)
		}
	}
}

func (l *IPLimitStore) cleanup(deadtime time.Duration) {
	now := time.Now()
	deadIPs := make([]string, 0, 20)

	l.mu.RLock()
	for ip, limiter := range l.limiters {
		limiter.mu.RLock()
		isDead := now.Sub(limiter.lastUpdate) >= deadtime
		limiter.mu.RUnlock()

		if isDead {
			deadIPs = append(deadIPs, ip)
		}
	}
	l.mu.RUnlock()

	if len(deadIPs) > 0 {
		l.mu.Lock()
		for _, ip := range deadIPs {
			delete(l.limiters, ip)
		}
		l.mu.Unlock()
	}
}

type ipLimiter struct {
	ip         string
	lastUpdate time.Time
}

func (l *IPLimitStore) reduce() {
	reduceIPcount := l.maxIP / 3
	if reduceIPcount <= 0 {
		reduceIPcount = 1
	}
	IPs := make([]ipLimiter, 0, l.maxIP)

	l.mu.RLock()
	for ip, lmt := range l.limiters {
		lmt.mu.RLock()
		lastUpdate := lmt.lastUpdate
		lmt.mu.RUnlock()
		IPs = append(IPs, ipLimiter{
			ip:         ip,
			lastUpdate: lastUpdate,
		})
	}
	l.mu.RUnlock()

	sort.Slice(IPs, func(l, r int) bool {
		return IPs[l].lastUpdate.Before(IPs[r].lastUpdate)
	})

	l.mu.Lock()
	for i := 0; i < reduceIPcount && i < len(IPs); i++ {
		delete(l.limiters, IPs[i].ip)
	}
	l.mu.Unlock()

	l.reducing.Store(false)
}
