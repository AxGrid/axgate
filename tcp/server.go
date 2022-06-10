package tcp

import (
	"errors"
	"fmt"
	pproto "github.com/axgrid/axgate/proto"
	bit_utils "github.com/axgrid/axgate/shared/bit-utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"
	"time"
)

var (
	services      = map[string]*GateConn{}
	servicesLock  = sync.Mutex{}
	connectionTTL = time.Second * 30
)

type GateConn struct {
	net.Conn
	lock     sync.Mutex
	name     string
	requests map[uint64]chan *pproto.GateResponse
	log      zerolog.Logger
}

func Send(request *pproto.GateRequest) (chan *pproto.GateResponse, error) {
	servicesLock.Lock()
	defer servicesLock.Unlock()
	conn, ok := services[request.Name]
	if !ok {
		return nil, fmt.Errorf("service %s not found", request.Name)
	}
	b, err := proto.Marshal(&pproto.Packet{
		Requests: request,
	})
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(bit_utils.AddSize(b))
	if err != nil {
		return nil, err
	}
	conn.log.Debug().Uint64("id", request.Id).Msg("send request")
	conn.lock.Lock()
	defer conn.lock.Unlock()
	conn.requests[request.Id] = make(chan *pproto.GateResponse)
	return conn.requests[request.Id], nil
}

func NewServer(bindAddress string) error {
	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		return err
	}
	log.Info().Str("address", bindAddress).Msg("start tcp-server")
	listener(l)
	return nil
}

func listener(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error().Err(err).Msg("error accepting connection")
			break
		}
		log.Debug().Str("remote-addr", conn.RemoteAddr().String()).Msg("new connection")
		gc := &GateConn{
			Conn:     conn,
			lock:     sync.Mutex{},
			requests: map[uint64]chan *pproto.GateResponse{},
		}
		go connection(gc)
	}
}

func connection(conn *GateConn) {
	defer conn.Close()
	conn.log = log.With().Str("remote-addr", conn.RemoteAddr().String()).Logger()
	err := conn.SetReadDeadline(time.Now().Add(connectionTTL))
	if err != nil {
		conn.log.Error().Err(err).Msg("set timeout error")
		return
	}
	dataChannel := make(chan []byte)
	go func() {
		for {
			data, ok := <-dataChannel
			if !ok {
				conn.log.Debug().Msg("channel closed")
				return
			}
			var p pproto.Packet
			err = proto.Unmarshal(data, &p)
			if err != nil {
				conn.log.Error().Err(err).Msg("fail to unmarshal")
				conn.Close()
				return
			}
			go process(&p, conn)
		}
	}()
	err = readerTL(conn, dataChannel)
	if err != nil {
		conn.log.Error().Err(err).Msg("read error")
	}

	//for {
	//	b, err := readTL(conn)
	//	_ = conn.SetReadDeadline(time.Now().Add(connectionTTL))
	//	if err != nil {
	//		conn.log.Error().Err(err).Msg("read-tl error")
	//		return
	//	}
	//	var p pproto.Packet
	//	err = proto.Unmarshal(b, &p)
	//	if err != nil {
	//		conn.log.Error().Err(err).Msg("fail to unmarshal proto")
	//		return
	//	}
	//	go process(&p, conn)
	//}
}

func process(p *pproto.Packet, conn *GateConn) {
	switch {
	case p.Handshake != nil && conn.name == "":
		conn.name = p.Handshake.Service
		conn.log = conn.log.With().Str("service", conn.name).Logger()
		conn.log.Info().Msg("handshake")
		servicesLock.Lock()
		defer servicesLock.Unlock()
		old, ok := services[conn.name]
		if ok {
			old.Close()
		}
		services[conn.name] = conn
		break
	case p.Responses != nil && conn.name != "":
		id := p.Responses.Id
		conn.lock.Lock()
		defer conn.lock.Unlock()
		ch, ok := conn.requests[id]
		if !ok {
			conn.log.Warn().Msg("request not found")
			return
		}
		ch <- p.Responses
		break
	case p.Ping != nil:
		b, err := proto.Marshal(&pproto.Packet{
			Pong: p.Ping,
		})
		log.Debug().Int64("ping", p.Ping.Time).Msg("ping")
		if err != nil {
			conn.Close()
		}
		conn.Write(bit_utils.AddSize(b))
		break
	}
}

func readTL(conn net.Conn) ([]byte, error) {
	sizeBuf := make([]byte, 4)
	sizeLen, err := conn.Read(sizeBuf)
	if err != nil {
		return nil, err
	}
	if sizeLen != 4 {
		return nil, errors.New("wrong size read length")
	}

	size := bit_utils.GetUInt32FromBytes(sizeBuf)
	log.Info().Hex("hex", sizeBuf).Uint32("size", size).Msg("read size")
	dataBuf := make([]byte, size)
	dataLen, err := conn.Read(dataBuf)
	if err != nil {
		return nil, err
	}

	if uint32(dataLen) != size {
		return nil, fmt.Errorf("wrong data read length. wait:%d read:%d", size, uint32(dataLen))
	}
	return dataBuf, nil
}

func readerTL(conn net.Conn, dataChannel chan []byte) error {
	defer conn.Close()
	buf := make([]byte, 4096)
	var data []byte
	for {
		i, err := conn.Read(buf)

		if err != nil {
			log.Error().Err(err).Msg("fail to read")
			return err
		}
		if i == 0 {
			log.Error().Err(err).Msg("connection closed")
			return errors.New("connection closed")
		}
		_ = conn.SetReadDeadline(time.Now().Add(connectionTTL))
		data = append(data, buf[:i]...)
		for { // Нужен если получили 2-ва пакета вместе
			ld := len(data)
			if ld >= 4 {
				l4 := bit_utils.GetUInt32FromBytes(data[:4])
				if uint32(ld) >= l4+4 {
					dataChannel <- data[4 : l4+4]
					data = data[l4+4:]
				} else {
					break
				}
			} else {
				break
			}
		}
	}
}
