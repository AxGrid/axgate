package main

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// -build-me-for: native
// -build-me-for: linux

var (
	port    int
	uri     string
	tcpHost string
	tcpPort int
)

func init() {
	flag.IntVar(&port, "port", 8081, "setup server port")
	flag.StringVar(&uri, "hosts", "localhost:8081", "set hosts, (,)separate")
	flag.StringVar(&tcpHost, "tcp-host", "0.0.0.0", "set tcp server host")
	flag.IntVar(&tcpPort, "tcp-port", 9090, "set tcp server port")
	flag.Parse()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05,000"}).Level(zerolog.DebugLevel)

}
