package sse

import "sync"

type Broadcaster struct {
	mu      sync.RWMutex
	clients map[string][]chan string // sessionID -> list of channels
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string][]chan string),
	}
}

func (b *Broadcaster) Register(sessionID string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[sessionID] = append(b.clients[sessionID], ch)
}

func (b *Broadcaster) Unregister(sessionID string, ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	channels := b.clients[sessionID]
	for i, c := range channels {
		if c == ch {
			b.clients[sessionID] = append(channels[:i], channels[i+1:]...)
			close(c)
			break
		}
	}
}

func (b *Broadcaster) Broadcast(sessionID string, msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.clients[sessionID] {
		select {
		case ch <- msg:
		default:
		}
	}
}
