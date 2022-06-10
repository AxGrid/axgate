package main

import (
	"github.com/axgrid/axgate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level := zerolog.InfoLevel
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05,000"}).Level(level)
	axgate.NewHTTPClient("bad", "route.axgrid.com:9090", "http://ya.ru/")
}
