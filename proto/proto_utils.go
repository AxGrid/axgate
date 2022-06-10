package proto

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func toGateHeader(header http.Header) []*GateHeader {
	var res []*GateHeader
	for k, v := range header {
		res = append(res, &GateHeader{
			Key:    k,
			Values: v,
		})
	}
	return res
}

func fromGateHeader(header []*GateHeader) http.Header {
	res := http.Header{}
	for _, h := range header {
		res[h.Key] = h.Values
	}
	return res
}

func (x *GateRequest) ToHttp() (*http.Request, error) {
	res, err := http.NewRequest(x.Method, x.Url, bytes.NewReader(x.Body))
	if err != nil {
		return nil, err
	}
	res.Header = fromGateHeader(x.Header)
	return res, nil
}
func (x *GateResponse) ToHttp(w http.ResponseWriter) error {
	for _, hd := range x.Header {
		w.Header().Add(hd.Key, strings.Join(hd.Values, ","))
	}
	w.Header().Add("x-gate-ref", x.Name)
	w.WriteHeader((int)(x.StatusCode))
	_, err := w.Write(x.Body)
	return err
}

func NewGateRequest(req *http.Request) (*GateRequest, error) {
	res := &GateRequest{
		Method:        req.Method,
		Url:           req.RequestURI,
		Header:        toGateHeader(req.Header),
		Host:          req.Host,
		RemoteAddr:    req.RemoteAddr,
		ContentLength: req.ContentLength,
	}

	switch req.Method {
	case "POST", "PUT", "PATCH":
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		res.Body = bodyBytes
	}
	return res, nil
}
func NewGateResponse(resp *http.Response) (*GateResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &GateResponse{
		StatusCode:    int32(resp.StatusCode),
		ContentLength: resp.ContentLength,
		Header:        toGateHeader(resp.Header),
		Body:          body,
	}
	return res, nil
}
