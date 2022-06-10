package main

import (
	"flag"
	"github.com/axgrid/axgate/handler"
	"github.com/axgrid/axgate/tcp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

// -build-me-for: native
// -build-me-for: linux

var (
	httpAddress string
	uri         string
	tcpAddress  string
	verbose     bool
)

func init() {
	flag.StringVar(&httpAddress, "http", ":8081", "setup http bind address")
	flag.StringVar(&uri, "hosts", "localhost:8081", "set http host names, (,)separate")
	flag.StringVar(&tcpAddress, "tcp", ":9090", "set tcp bind address :9090")
	flag.BoolVar(&verbose, "verbose", false, "show more debug lines")
	flag.Parse()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05,000"}).Level(level)
	go func() {
		err := tcp.NewServer(tcpAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("fail to start tcp server")
		}
	}()
	err := handler.NewHandler(httpAddress, strings.Split(uri, ","), verbose)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to start http-listener")
	}
}
