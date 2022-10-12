package main

import (
	"fmt"
	"sync"
	"time"
)

// Publish/subscribe server

type Event string

type PubSub interface {
	// Publish publishes the event e to
	// all current subscriptions.
	Publish(e Event)
	// Subscribe registers c to receive future events.
	// All subscribers receive events in the same order,
	// and that order respects program order:
	// if Publish(e1) happens before Publish(e2),
	// subscribers receive e1 before e2.
	Subscribe(c chan<- Event)
	// Cancel cancels the prior subscription of channel c.
	// After any pending already-published events
	// have been sent on c, the server will signal that the
	// subscription is cancelled by closing c.
	Cancel(c chan<- Event)
}

///////// Straightforward implementation with mutex
type ServerMutex struct {
	mu  sync.Mutex
	sub map[chan<- Event]bool
}

func (s *ServerMutex) Init() {
	s.sub = make(map[chan<- Event]bool)
}
func (s *ServerMutex) Publish(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.sub {
		c <- e
	}
}
func (s *ServerMutex) Subscribe(c chan<- Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sub[c] {
		panic("pubsub: already subscribed")
	}
	s.sub[c] = true
}
func (s *ServerMutex) Cancel(c chan<- Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.sub[c] {
		panic("pubsub: not subscribed")
	}
	close(c)
	delete(s.sub, c)
}

/////////
type ServerChannels struct {
	publish   chan Event
	subscribe chan subReq
	cancel    chan subReq
}
type subReq struct {
	c  chan<- Event
	ok chan bool
}

func (s *ServerChannels) Init() {
	s.publish = make(chan Event)
	s.subscribe = make(chan subReq)
	s.cancel = make(chan subReq)
	go s.loop()
}

func (s *ServerChannels) Publish(e Event) {
	s.publish <- e
}

func (s *ServerChannels) Subscribe(c chan<- Event) {
	r := subReq{c: c, ok: make(chan bool)}
	s.subscribe <- r
	if !<-r.ok {
		panic("pubsub: already subscribed")
	}
}

func (s *ServerChannels) Cancel(c chan<- Event) {
	r := subReq{c: c, ok: make(chan bool)}
	s.cancel <- r
	if !<-r.ok {
		panic("pubsub: not subscribed")
	}
}

func (s *ServerChannels) loop() {
	sub := make(map[chan<- Event]bool)
	for {
		select {
		case e := <-s.publish:
			for c := range sub {
				c <- e
			}
		case r := <-s.subscribe:
			if sub[r.c] {
				r.ok <- false
				break
			}
			sub[r.c] = true
			r.ok <- true
		case r := <-s.cancel:
			if !sub[r.c] {
				r.ok <- false
				break
			}
			close(r.c)
			delete(sub, r.c)
			r.ok <- true
		}
	}
}

/////////
type ServerMutexHelper struct {
	mu  sync.Mutex
	sub map[chan<- Event]chan<- Event
}

func (s *ServerMutexHelper) Init() {
	s.sub = make(map[chan<- Event]chan<- Event)
}

func (s *ServerMutexHelper) Publish(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, h := range s.sub {
		h <- e
	}
}

func (s *ServerMutexHelper) Subscribe(c chan<- Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sub[c] != nil {
		panic("pubsub: already subscribed")
	}
	h := make(chan Event)
	go helper(h, c)
	s.sub[c] = h
}

func (s *ServerMutexHelper) Cancel(c chan<- Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.sub[c]
	if !ok {
		panic("pubsub: not subscribed")
	}

	close(h)
	delete(s.sub, c)
}

////////////////////////////
type ServerChannelsHelper struct {
	publish   chan Event
	subscribe chan subReq
	cancel    chan subReq
}

func (s *ServerChannelsHelper) Init() {
	s.publish = make(chan Event)
	s.subscribe = make(chan subReq)
	s.cancel = make(chan subReq)
	go s.loop()
}

func (s *ServerChannelsHelper) Publish(e Event) {
	s.publish <- e
}

func (s *ServerChannelsHelper) Subscribe(c chan<- Event) {
	r := subReq{c: c, ok: make(chan bool)}
	s.subscribe <- r
	if !<-r.ok {
		panic("pubsub: already subscribed")
	}
}

func (s *ServerChannelsHelper) Cancel(c chan<- Event) {
	r := subReq{c: c, ok: make(chan bool)}
	s.cancel <- r
	if !<-r.ok {
		panic("pubsub: not subscribed")
	}
}

func (s *ServerChannelsHelper) loop() {
	sub := make(map[chan<- Event]chan<- Event)
	for {
		select {
		case e := <-s.publish:
			for _, h := range sub {
				h <- e
			}
		case r := <-s.subscribe:
			if sub[r.c] != nil {
				r.ok <- false
				break
			}
			h := make(chan Event)
			go helper(h, r.c)
			sub[r.c] = h
			r.ok <- true
		case r := <-s.cancel:
			if sub[r.c] == nil {
				r.ok <- false
				break
			}
			close(sub[r.c])
			delete(sub, r.c)
			r.ok <- true
		}
	}
}

func helper(in <-chan Event, out chan<- Event) {
	var q []Event

	for in != nil || (in == nil && len(q) > 0) {
		// Decide whether and what to send.
		var (
			next    Event
			sendOut chan<- Event
		)

		if len(q) > 0 {
			sendOut = out
			next = q[0]
		}

		select {
		case e, ok := <-in:
			if !ok {
				in = nil // stop receiving from in
				break
			}
			q = append(q, e)
		case sendOut <- next:
			if len(q) != 0 {
				q = q[1:]
			}
		}
	}

	close(out)
}

func test(name string, s PubSub, subCount int, msgCount int) {
	ts := time.Now()
	subs := make([]chan Event, 0, subCount)
	done := make(chan bool)

	for i := 0; i < subCount; i++ {
		subs = append(subs, make(chan Event))
		s.Subscribe(subs[i])
	}

	for i := 0; i < subCount; i++ {
		go func(i int) {
			for e := range subs[i] {
				_ = e
				// fmt.Printf("%d: %v\n", i, e)
			}
			done <- true
		}(i)
	}

	go func() {
		for i := 0; i < msgCount; i++ {
			s.Publish(Event(fmt.Sprintf("event%d", i+1)))
		}

		for i := 0; i < subCount; i++ {
			s.Cancel(subs[i])
		}
	}()

	for i := 0; i < subCount; i++ {
		<-done
	}

	fmt.Printf("%s \t time: %v\n", name, time.Since(ts))
}

func pattern1() {
	// Publish/subscribe server

	cSubs := 1000
	cMsgs := 100000

	s1 := &ServerMutex{}
	s1.Init()
	test("mutex", s1, cSubs, cMsgs)

	s3 := &ServerChannels{}
	s3.Init()
	test("chs", s3, cSubs, cMsgs)

	s2 := &ServerMutexHelper{}
	s2.Init()
	test("mu+q", s2, cSubs, cMsgs)

	s4 := &ServerChannelsHelper{}
	s4.Init()
	test("chs+q", s4, cSubs, cMsgs)
}
