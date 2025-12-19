package main

import (
	"context"
	"log/slog"
	protov1 "marmoset/gen/proto/v1"
	"marmoset/gen/proto/v1/protov1connect"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/google/uuid"
	"github.com/lmittmann/tint"
)

type GreetServer struct{}

func (s *GreetServer) Greet(
	_ context.Context,
	req *protov1.GreetMsg,
) (*protov1.GreetMsg, error) {
	res := &protov1.GreetMsg{
		Id:  uuid.New().String(),
		Uri: "http://localhost:9999",
	}
	return res, nil
}

func main() {
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))

	greeter := &GreetServer{}
	mux := http.NewServeMux()
	path, handler := protov1connect.NewGossipServiceHandler(
		greeter,
		// Validation via Protovalidate is almost always recommended
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	mux.Handle(path, handler)
	p := new(http.Protocols)
	p.SetHTTP2(true)
	// Use h2c so we can serve HTTP/2 without TLS.
	p.SetUnencryptedHTTP2(true)
	s := http.Server{
		Addr:      "localhost:8080",
		Handler:   mux,
		Protocols: p,
	}
	s.ListenAndServe()
}
