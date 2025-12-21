package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"marmoset/gen/proto/v1/protov1connect"
	"marmoset/src/cluster"
	"marmoset/src/state"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/lmittmann/tint"
)

func main() {
	uri := flag.String("uri", "", "uri")
	listen := flag.String("listen", "", "listen")
	contact := flag.String("contact", "", "contact")
	flag.Parse()

	// todo validation blah blah blah
	if strings.Compare("", *uri) == 0 || strings.Compare("", *listen) == 0 {
		panic("wrong args")
	}

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))

	clusterServer := cluster.NewClusterMgr(*uri, *contact)
	state.NewState(clusterServer)

	mux := http.NewServeMux()
	path, handler := protov1connect.NewClusterServiceHandler(
		clusterServer,
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	mux.Handle(path, handler)

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	s := http.Server{
		Addr:      *listen,
		Handler:   mux,
		Protocols: p,
	}

	go s.ListenAndServe()
	slog.Info(fmt.Sprintf("Listening on: %s - uri: %s", *listen, *uri))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		// temporary debugging repl
		fmt.Print("> ")
		if !scanner.Scan() {
			break // EOF or error
		}
		line := strings.TrimSpace(scanner.Text())

		if line == "exit" {
			fmt.Println("Goodbye!")
			break
		} else if line == "peers" {
			clusterServer.ListPeers()
		} else if line == "cinc" {
			clusterServer.IncCounter()
		} else {
			// process input
			fmt.Println("VALID CMDS: exit peers cinc")
		}
	}
}
