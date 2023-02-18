package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type event struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type subscriber struct {
	conn  chan<- event
	topic string
}

type broker struct {
	subscribers map[string][]*subscriber
	mux         sync.RWMutex
}

func (b *broker) subscribe(topic string, conn chan<- event) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], &subscriber{conn: conn, topic: topic})
}

func (b *broker) unsubscribe(topic string, conn chan<- event) {
	b.mux.Lock()
	defer b.mux.Unlock()

	subscribers, ok := b.subscribers[topic]
	if !ok {
		return
	}

	for i, sub := range subscribers {
		if sub.conn == conn {
			close(sub.conn)
			b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
			return
		}
	}
}

func (b *broker) publish(ev event) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	subscribers, ok := b.subscribers[ev.Topic]
	if !ok {
		return
	}

	for _, sub := range subscribers {
		sub.conn <- ev
	}
}

func main() {
	b := &broker{
		subscribers: make(map[string][]*subscriber),
	}

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")

		conn, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		events := make(chan event)

		b.subscribe(topic, events)

		defer func() {
			b.unsubscribe(topic, events)
		}()

		for {
			select {
			case ev, ok := <-events:
				if !ok {
					return
				}
				msg, _ := json.Marshal(ev)
				fmt.Fprintf(w, "topic: %s\ndata: %s\n\n", topic, msg)
				conn.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		message := r.URL.Query().Get("message")

		ev := event{
			Topic:   topic,
			Message: message,
		}

		b.publish(ev)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
