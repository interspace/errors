package memlistener

import "net"
import "errors"
import "sync/atomic"

type MemoryListener struct {
	connections   chan net.Conn
	state         chan int
	isStateClosed uint32
}

func NewMemoryListener() *MemoryListener {
	ml := &MemoryListener{}
	ml.connections = make(chan net.Conn)
	ml.state = make(chan int)
	return ml
}

func (ml *MemoryListener) Accept() (net.Conn, error) {
	select {
	case newConnection := <-ml.connections:
		return newConnection, nil
	case <-ml.state:
		return nil, errors.New("Listener closed")
	}
}

func (ml *MemoryListener) Close() error {
	if atomic.CompareAndSwapUint32(&ml.isStateClosed, 0, 1) {
		close(ml.state)
	}
	return nil
}

func (ml *MemoryListener) Dial(network, addr string) (net.Conn, error) {
	select {
	case <-ml.state:
		return nil, errors.New("Listener closed")
	default:
	}
	//Create an in memory transport
	serverSide, clientSide := net.Pipe()
	//Pass half to the server
	ml.connections <- serverSide
	//Return the other half to the client
	return clientSide, nil

}

type memoryAddr int

func (memoryAddr) Network() string {
	return "memory"
}

func (memoryAddr) String() string {
	return "local"
}
func (ml *MemoryListener) Addr() net.Addr {
	return memoryAddr(0)
}
