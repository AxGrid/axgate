package handler

import (
	"fmt"
	pproto "github.com/axgrid/axgate/proto"
	"github.com/axgrid/axgate/tcp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"strings"
)

func NewHandler(httpAddress string, hosts []string, verbose bool) error {
	var stringHost string
	if len(hosts) == 1 {
		stringHost = regexp.QuoteMeta(hosts[0])
	} else {
		var qHost []string
		for _, h := range hosts {
			qHost = append(qHost, regexp.QuoteMeta(h))
		}

		stringHost = fmt.Sprintf("(%s)", strings.Join(qHost, "|"))
	}
	rstr := "^(?P<service>[A-z0-9_-]+)\\." + stringHost
	hostMatcher, err := regexp.Compile(rstr)
	if err != nil {
		return err
	}
	r := chi.NewRouter()
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.ErrorLevel
	}
	httpLogger := log.With().Str("service", "http").Logger().Level(level)
	r.Use(httplog.RequestLogger(httpLogger))
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		matches := hostMatcher.FindStringSubmatch(r.Host)
		if len(matches) == 0 {
			root(w, r)
		} else {
			err := service(matches[1], w, r)
			if err != nil {
				w.WriteHeader(500)
				w.Write(([]byte)("500 internal server error: " + err.Error()))
			}
		}
	})
	log.Info().Str("address", httpAddress).Msg("start http-listener")
	return http.ListenAndServe(httpAddress, r)
}

func root(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("ROOT!")
	w.Header().Set("x-gate", "root")
	w.Write(([]byte)("ROOT"))
}

func service(name string, w http.ResponseWriter, r *http.Request) (err error) {
	rq, err := pproto.NewGateRequest(r)
	rq.Name = name
	if err != nil {
		return err
	}
	c, err := tcp.Send(rq)
	if err != nil {
		return err
	}
	rs := <-c
	return rs.ToHttp(w)
}
