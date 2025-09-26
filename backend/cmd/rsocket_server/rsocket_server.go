package main

import (
	"context"
	"log"
	"time"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
)

func main() {
	// Server listen di TCP port 7878
	err := rsocket.Receive().
		Acceptor(func(ctx context.Context, setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			return rsocket.NewAbstractSocket(
				// Handle Request-Response
				rsocket.RequestResponse(func(msg payload.Payload) mono.Mono {
					meta, _ := msg.MetadataUTF8()
					log.Printf("RR <= data=%q meta=%q", msg.DataUTF8(), meta)
					// Balikin response "pong:<data>"
					return mono.Just(payload.NewString("pong:"+msg.DataUTF8(), time.Now().Format(time.RFC3339Nano)))
				}),
			), nil
		}).
		Transport(rsocket.TCPServer().SetAddr(":7878").Build()).
		Serve(context.Background())

	log.Fatalln(err)
}
