package handler

import (
	"bytes"
	_ "embed"
	"fmt"
	pproto "github.com/axgrid/axgate/proto"
	"github.com/axgrid/axgate/tcp"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

//go:embed "template/index.gohtml"
var index []byte

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
			root(w, r, hosts[0])
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

func root(w http.ResponseWriter, r *http.Request, host string) {
	var res []*Info
	for _, name := range tcp.GetServicesNames() {
		res = append(res, &Info{
			Name: name,
			Url:  fmt.Sprintf("http://%s.%s", name, host),
		})
	}

	b, err := render(index, res)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
	w.WriteHeader(200)
	w.Write(b)
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

func render(templateByte []byte, data interface{}) ([]byte, error) {
	t, err := template.New("").Parse(string(templateByte))
	if err != nil {
		log.Error().Err(err).Msg("failed to create template")
		return nil, err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to render template")
		return nil, err
	}
	return tpl.Bytes(), nil
}

type Info struct {
	Name string
	Url  string
}
