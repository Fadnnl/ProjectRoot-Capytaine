package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_PAYLOAD_SIZE_KB = 1
	ALPHANUMERIC_CHARS      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	MAX_PAYLOAD_SIZE_KB     = 1024 * 10 // 10 MB
)

func generateRandomPayload(sizeKB int) string {
	totalChars := sizeKB * 1024
	var b strings.Builder
	b.Grow(totalChars)
	for i := 0; i < totalChars; i++ {
		b.WriteByte(ALPHANUMERIC_CHARS[rand.Intn(len(ALPHANUMERIC_CHARS))])
	}
	return b.String()
}

func atoiOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// non-stream
func helloHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format(time.RFC3339Nano)
	sizeKB := clamp(atoiOrDefault(r.URL.Query().Get("size"), DEFAULT_PAYLOAD_SIZE_KB), 1, MAX_PAYLOAD_SIZE_KB)
	payload := generateRandomPayload(sizeKB)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hello! Request at %s\nPayload size: %d KB\nPayload: %s\n", now, sizeKB, payload)
}

// SSE streaming
func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sizeKB := clamp(atoiOrDefault(r.URL.Query().Get("size"), DEFAULT_PAYLOAD_SIZE_KB), 1, MAX_PAYLOAD_SIZE_KB)
	interval := time.Duration(atoiOrDefault(r.URL.Query().Get("interval_ms"), 1000)) * time.Millisecond
	count := atoiOrDefault(r.URL.Query().Get("count"), 10)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := r.Context()
	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("SSE client disconnected")
			return
		case t := <-ticker.C:
			i++
			event := map[string]any{
				"ts":      t.Format(time.RFC3339Nano),
				"idx":     i,
				"size_kb": sizeKB,
				"payload": generateRandomPayload(sizeKB),
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: tick\ndata: %s\n\n", data)
			flusher.Flush()
			if count > 0 && i >= count {
				return
			}
		}
	}
}

// chunked streaming
func chunkHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	sizeKB := clamp(atoiOrDefault(r.URL.Query().Get("size"), DEFAULT_PAYLOAD_SIZE_KB), 1, MAX_PAYLOAD_SIZE_KB)
	interval := time.Duration(atoiOrDefault(r.URL.Query().Get("interval_ms"), 1000)) * time.Millisecond
	count := atoiOrDefault(r.URL.Query().Get("count"), 10)

	w.Header().Set("Content-Type", "application/octet-stream")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := r.Context()
	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Chunk client disconnected")
			return
		case t := <-ticker.C:
			i++
			chunk := fmt.Sprintf("idx=%d ts=%s size_kb=%d payload=%s\n",
				i, t.Format(time.RFC3339Nano), sizeKB, generateRandomPayload(sizeKB))
			_, _ = io.WriteString(w, chunk)
			flusher.Flush()
			if count > 0 && i >= count {
				return
			}
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.WriteString(w, "ok")
}

func payloadHandler(w http.ResponseWriter, r *http.Request) {
    sizeKB := clamp(atoiOrDefault(r.URL.Query().Get("size"), DEFAULT_PAYLOAD_SIZE_KB), 1, MAX_PAYLOAD_SIZE_KB)
    payload := generateRandomPayload(sizeKB)

    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("X-Payload-Size", fmt.Sprintf("%d KB", sizeKB))
    _, _ = io.WriteString(w, payload)
}


func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hello/stream", sseHandler)
	http.HandleFunc("/hello/chunk", chunkHandler)
	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/payload", payloadHandler)
	log.Println("Backend running on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
