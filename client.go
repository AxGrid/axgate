package axgate

import (
	"bytes"
	"errors"
	"fmt"
	pproto "github.com/axgrid/axgate/proto"
	"github.com/axgrid/axgate/tcp"
	"net/http"
	"strings"
	"time"
)

var tr = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    time.Second * 20,
	DisableCompression: true,
}

func NewHTTPHandlerClient(name string, gateAddress string, handler http.Handler, args ...string) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	return tcp.NewClient(name, gateAddress, func(request *pproto.GateRequest) (*pproto.GateResponse, error) {
		wr := &ResponseWriter{
			header: http.Header{},
			code:   200,
		}
		hr, err := request.ToHttp()
		if err != nil {
			return nil, err
		}
		handler.ServeHTTP(wr, hr)
		return wr.ToGate()
	}, args...)
}

func NewHTTPClient(name string, gateAddress string, requestAddress string, args ...string) error {
	client := &http.Client{Transport: tr}
	if strings.HasSuffix(requestAddress, "/") {
		requestAddress = requestAddress[:len(requestAddress)-1]
	}
	return tcp.NewClient(name, gateAddress, func(request *pproto.GateRequest) (*pproto.GateResponse, error) {
		httpRequest, err := http.NewRequest(request.Method, fmt.Sprintf("%s%s", requestAddress, request.Url), bytes.NewReader(request.Body))
		if err != nil {
			return nil, err
		}
		httpRequest.Header = pproto.FromGateHeader(request.Header)
		httpResponse, err := client.Do(httpRequest)
		if err != nil {
			return nil, err
		}
		resp, err := pproto.NewGateResponse(httpResponse)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}, args...)
}

type ResponseWriter struct {
	body   []byte
	code   int
	header http.Header
}

func (c *ResponseWriter) Header() http.Header { return c.header }

func (c *ResponseWriter) Write(data []byte) (int, error) {
	c.body = append(c.body, data...)
	return len(data), nil
}

func (c *ResponseWriter) WriteHeader(statusCode int) {
	c.code = statusCode
}

func (c *ResponseWriter) ToGate() (*pproto.GateResponse, error) {
	res := &pproto.GateResponse{
		StatusCode:    int32(c.code),
		ContentLength: int64(len(c.body)),
		Header:        pproto.ToGateHeader(c.header),
		Body:          c.body,
	}
	return res, nil
}
