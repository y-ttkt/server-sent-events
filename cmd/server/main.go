package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/y-ttkt/server-sent-events/internal/sse"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type tick struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/stream", stream)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%s", getenv("PORT", "8080")),
		Handler:           logMiddleware(corsMiddleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// グレースフルシャットダウン
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("bye")
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			allowOrigin := os.Getenv("ALLOW_ORIGIN")
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	lastID := r.Header.Get("Last-Event-ID")
	if lastID != "" {
		log.Printf("client resume from id=%s", lastID)
	}

	retryMS := 5000
	if q := r.URL.Query().Get("retry"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			retryMS = v
		}
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	send := time.NewTicker(1 * time.Second)
	defer send.Stop()

	id := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			_ = sse.Heartbeat(w, "keep-alive")
			flusher.Flush()
		case <-send.C:
			id++
			ev := sse.Event{
				Name:  "tick",
				ID:    strconv.Itoa(id),
				Retry: retryMS,
				Data: tick{
					Time:  time.Now().Format(time.RFC3339Nano),
					Value: 90 + rand.Float64()*20,
				},
			}
			if err := ev.WriteTo(w); err != nil {
				// client likely went away
				log.Printf("write error: %v", err)
				return
			}
			flusher.Flush()
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
