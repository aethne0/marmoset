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
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/lmittmann/tint"
)

func main() {
	uri := flag.String("uri", "", "uri")
	pprof := flag.String("pprof", "", "pprof")
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
	stateMgr := state.NewStateMgr(clusterServer)

	mux := http.NewServeMux()
	path, handler := protov1connect.NewClusterServiceHandler(
		clusterServer,
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	mux.Handle(path, handler)
	pathState, handlerState := protov1connect.NewStateServiceHandler(
		stateMgr,
		connect.WithInterceptors(validate.NewInterceptor()),
	)
	mux.Handle(pathState, handlerState)

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	s := http.Server{
		Addr:      *listen,
		Handler:   mux,
		Protocols: p,
	}

	go http.ListenAndServe(*pprof, nil)
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
			clusterServer.PrintListPeers()
		} else if line == "cinc" {
			clusterServer.IncCounter()
		} else if line == "orset" {
			stateMgr.PrintORSet()
		} else if line == "vec" {
			stateMgr.PrintVector()
		} else if line == "pvec" {
			stateMgr.PrintPeerVectors()
		} else if strings.Split(line, " ")[0] == "put" {
			k := strings.Split(line, " ")[1]
			stateMgr.SetInsert(k)
			fmt.Printf("put: %s\n", k)
		} else if strings.Split(line, " ")[0] == "del" {
			k := strings.Split(line, " ")[1]
			stateMgr.SetRemove(k)
			fmt.Printf("del: %s\n", k)
		} else if strings.Split(line, " ")[0] == "get" {
			k := strings.Split(line, " ")[1]
			has := stateMgr.SetHas(k)
			if has {
				fmt.Printf("get %s: TRUE\n", k)
			} else {
				fmt.Printf("get %s: FALSE\n", k)
			}
		} else {
			// process input
			fmt.Println("VALID CMDS:\nexit, peers, cinc, put [key], get [key], orset, vec, pvec")
		}
	}
}
