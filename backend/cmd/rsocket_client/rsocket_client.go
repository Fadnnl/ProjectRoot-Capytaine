package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	host := getenv("RSOCKET_HOST", "127.0.0.1")
	portStr := getenv("RSOCKET_PORT", "7878")
	port, _ := strconv.Atoi(portStr)

	cli, err := rsocket.Connect().
		SetupPayload(payload.NewString("setup", "from-go")).
		Transport(rsocket.TCPClient().SetHostAndPort(host, port).Build()).
		Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// Kirim request-response
	resp, err := cli.RequestResponse(payload.NewString("ping", time.Now().Format(time.RFC3339Nano))).Block(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	meta, _ := resp.MetadataUTF8()
	log.Printf("RR => response data=%q meta=%q", resp.DataUTF8(), meta)
}
