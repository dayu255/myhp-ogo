package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/dayu255/myhp-ogo/src/limit"
	"github.com/dayu255/myhp-ogo/src/og"
	"golang.org/x/time/rate"
)

const LISTEN_PORT = 8080

func main() {
	mux := http.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limiter := limit.NewLimitStore(
		rate.Every(3*time.Second),
		10,
		int(math.Pow10(5)),
	)
	go limiter.StartCleanup(ctx, 1*time.Hour, 1*time.Hour)

	mux.HandleFunc("/og.png",
		func(w http.ResponseWriter, r *http.Request) {
			// Method Check
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// IP limiter
			ip, err := getIP(r)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !limiter.Allow(ip) {
				http.Error(w, "Too much request", http.StatusTooManyRequests)
				return
			}

			og.Og(w, r)
		},
	)

	log.Printf("Listening PORT: %d\n", LISTEN_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", LISTEN_PORT), mux)
}

func getIP(r *http.Request) (ip string, err error) {
	if ip = r.Header.Get("CF-Connecting-IP"); ip == "" {
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", err
		}
	}
	return ip, nil
}
