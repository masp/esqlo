package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/masp/esqlo/esqlo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	addr     = flag.String("l", "127.0.0.1:8080", "address to listen on")
	verbose  = flag.Bool("v", false, "verbose?")
	serveDir = flag.String("s", "", "serve a directory of templates")
)

func init() {
	flag.StringVar(serveDir, "serve", "", "serve a directory of templates")
	flag.StringVar(addr, "listen", "127.0.0.1:8080", "address to listen on")
}

func main() {
	flag.Parse()
	if *serveDir == "" {
		fmt.Fprintf(os.Stderr, "usage: %s -s <directory of html files>", os.Args[0])
		os.Exit(1)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ll := zerolog.InfoLevel
	if *verbose {
		ll = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(ll)

	h := esqlo.RenderAll(http.FileServer(http.Dir(*serveDir)))
	http.Handle("/", h)
	log.Info().Msgf("listening on %s", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: listen and serve: %v", err)
	}
}
