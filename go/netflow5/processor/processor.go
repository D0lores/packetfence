package processor

import (
	"errors"
	"github.com/inverse-inc/packetfence/go/bytearraypool"
	"github.com/inverse-inc/packetfence/go/bytesdispatcher"
	"github.com/inverse-inc/packetfence/go/netflow5"
	"net"
	"runtime"
	"strings"
	"unsafe"
)

type Handler func(header *netflow5.Header, i int, flow *netflow5.Flow)

type Processor struct {
	Conn              net.PacketConn
	Handler           Handler
	Workers           int
	BacklogSize       int
	MaxPacketSize     int
	ByteArrayPoolSize int
	byteArrayPool     *bytearraypool.ByteArrayPool
	stopChan          chan struct{}
}

func (p *Processor) setDefaults() {
	if p.Handler == nil {
		panic(errors.New("No handler defined"))
	}

	if p.Workers <= 0 {
		p.Workers = runtime.GOMAXPROCS(0)
	}

	if p.MaxPacketSize <= 0 {
		p.MaxPacketSize = 2048
	}

	if p.BacklogSize <= 0 {
		p.BacklogSize = 100
	}

	if p.ByteArrayPoolSize <= 0 {
		p.ByteArrayPoolSize = p.Workers * 2
	}

	p.byteArrayPool = bytearraypool.NewByteArrayPool(p.ByteArrayPoolSize, p.MaxPacketSize)

	if p.Conn == nil {
		conn, err := net.ListenPacket("udp", "127.0.0.1:2055")
		if err != nil {
			panic(err)
		}

		p.Conn = conn
	}

	if p.stopChan == nil {
		p.stopChan = make(chan struct{}, 1)
	}
}

func BytesHandlerForNetFlow5Handler(h Handler) bytesdispatcher.BytesHandler {
	return func(buffer []byte) {
		var data *netflow5.NetFlow5
		data = (*netflow5.NetFlow5)(unsafe.Pointer(&buffer[0]))
		count := data.Header.Length()
		for i := 0; i < int(count); i++ {
			h(&data.Header, i, &data.Flows[i])
		}
	}
}

func (p *Processor) dispatcher() *bytesdispatcher.Dispatcher {
	return bytesdispatcher.NewDispatcher(p.Workers, p.BacklogSize, BytesHandlerForNetFlow5Handler(p.Handler), p.byteArrayPool)
}

func (p *Processor) Stop() {
	c := p.stopChan
	p.stopChan = nil
	c <- struct{}{}
	p.Conn.Close()
}

func (p *Processor) isCloseError(err error) bool {
	if p.stopChan != nil {
		return false
	}

	str := err.Error()
	return strings.Contains(str, "use of closed network connection")
}

func (p *Processor) Start() {
	p.setDefaults()
	dispatcher := p.dispatcher()
	dispatcher.Run()
	stopChan := p.stopChan

LOOP:
	for {
		buffer := p.byteArrayPool.Get()
		rlen, remote, err := p.Conn.ReadFrom(buffer)
		if err != nil {
			if p.isCloseError(err) {
				break
			}

			panic(err)
		}
		_, _ = rlen, remote
		dispatcher.SubmitJob(buffer)
		select {
		case <-stopChan:
			break LOOP
		default:
			continue
		}
	}

	dispatcher.Stop()
}