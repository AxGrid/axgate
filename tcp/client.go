package tcp

import (
	"bytes"
	"errors"
	"fmt"
	pproto "github.com/axgrid/axgate/proto"
	bit_utils "github.com/axgrid/axgate/shared/bit-utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"net"
	"net/http"
	"strings"
	"time"
)

var reconnectTTL = time.Millisecond * 100
var pingTTL = time.Second * 10
var tr = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    time.Second * 20,
	DisableCompression: true,
}

type fListener func(request *pproto.GateRequest) (*pproto.GateResponse, error)

func NewClient(name string, gateAddress string, listener fListener) (err error) {
	log.Info().Str("name", name).Str("address", gateAddress).Msg("start gate-client")
	tcpAddr, err := net.ResolveTCPAddr("tcp", gateAddress)
	if err != nil {
		return err
	}
	for {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Debug().Err(err).Msg("fail to create tcp-connection")
			time.Sleep(reconnectTTL)
			continue
		}
		err = handshake(conn, name)
		if err != nil {
			log.Error().Err(err).Msg("fail to send handshake")
			continue
		}
		ex := ping(conn)
		err = clientLoop(conn, listener)
		ex <- true
		if err != nil {
			log.Error().Err(err).Msg("client error")
		}
	}
}

func ping(conn net.Conn) chan bool {
	pingInterval := time.NewTicker(pingTTL)
	closeChan := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-pingInterval.C:
				p := &pproto.Packet{
					Ping: &pproto.GatePing{
						Time: time.Now().UnixMilli(),
					},
				}
				b, err := proto.Marshal(p)
				if err != nil {
					return
				}
				_, err = conn.Write(bit_utils.AddSize(b))
				if err != nil {
					return
				}
			case <-closeChan:
				return
			}
		}
	}()
	return closeChan
}

func clientLoop(conn net.Conn, listener fListener) error {
	for {
		b, err := readTL(conn)
		if err != nil {
			return err
		}
		var p pproto.Packet
		err = proto.Unmarshal(b, &p)
		if err != nil {
			return err
		}
		switch {
		case p.Pong != nil:
			log.Debug().Int64("ms", time.Now().UnixMilli()-p.Pong.Time).Msg("ping")
			break
		case p.Requests != nil:
			go func(conn net.Conn) {
				resp, err := listener(p.Requests)
				if err != nil {
					log.Error().Err(err).Msg("error in listener")
					return
				}
				resp.Id = p.Requests.Id
				resp.Name = p.Requests.Name
				p := &pproto.Packet{
					Responses: resp,
				}
				b, err := proto.Marshal(p)
				if err != nil {
					log.Error().Err(err).Msg("error marshal response")
					return
				}
				_, err = conn.Write(bit_utils.AddSize(b))
				if err != nil {
					log.Error().Err(err).Msg("error send data")
					return
				}
			}(conn)
		}
	}
}

func handshake(conn net.Conn, name string) error {
	pck := &pproto.Packet{
		Handshake: &pproto.GateHandshake{
			Service: name,
		},
	}
	data, err := proto.Marshal(pck)
	if err != nil {
		return err
	}
	_, err = conn.Write(bit_utils.AddSize(data))
	if err != nil {
		return err
	}
	return nil
}

func NewHTTPHandlerClient(name string, gateAddress string, handler http.Handler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	return NewClient(name, gateAddress, func(request *pproto.GateRequest) (*pproto.GateResponse, error) {
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
	})
}

func NewHTTPClient(name string, gateAddress string, requestAddress string) error {
	client := &http.Client{Transport: tr}
	if strings.HasSuffix(requestAddress, "/") {
		requestAddress = requestAddress[:len(requestAddress)-1]
	}
	return NewClient(name, gateAddress, func(request *pproto.GateRequest) (*pproto.GateResponse, error) {
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
	})
}
