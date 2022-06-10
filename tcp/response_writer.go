package tcp

import (
	pproto "github.com/axgrid/axgate/proto"
	"net/http"
)

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
