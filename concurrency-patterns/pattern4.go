package main

import "sync"

// Protocol multiplexer

type ProtocolMux interface {
	// Init initializes the mux to manage messages to the given service.
	Init(Service)
	// Call makes a request with the given message and returns the reply.
	// Multiple goroutines may call Call concurrently.
	Call(Msg) Msg
}

type Service interface {
	// ReadTag returns the muxing identifier in the request or reply message.
	// Multiple goroutines may call ReadTag concurrently.
	ReadTag(Msg) int64
	// Send sends a request message to the remote service.
	// Send must not be called concurrently with itself.
	Send(Msg)
	// Recv waits for and returns a reply message from the remote service.
	// Recv must not be called concurrently with itself.
	Recv() Msg
}

type Msg string

type Mux struct {
	srv     Service
	send    chan Msg
	mu      sync.Mutex
	pending map[int64]chan<- Msg
}

func (m *Mux) Init(srv Service) {
	m.srv = srv
	m.pending = make(map[int64]chan<- Msg)

	go m.sendLoop()
	go m.recvLoop()
}

func (m *Mux) sendLoop() {
	for args := range m.send {
		m.srv.Send(args)
	}
}

func (m *Mux) recvLoop() {
	for {
		reply := m.srv.Recv()
		tag := m.srv.ReadTag(reply)

		m.mu.Lock()
		done := m.pending[tag]
		delete(m.pending, tag)
		m.mu.Unlock()

		if done == nil {
			panic("unexpected reply")
		}

		done <- reply
	}
}

func (m *Mux) Call(args Msg) (reply Msg) {
	tag := m.srv.ReadTag(args)
	done := make(chan Msg, 1)

	m.mu.Lock()
	if m.pending[tag] != nil {
		m.mu.Unlock()
		panic("mux: duplicate call tag")
	}

	m.pending[tag] = done
	m.mu.Unlock()

	m.send <- args

	return <-done
}
