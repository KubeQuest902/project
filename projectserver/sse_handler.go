package projectserver

import (
	"fmt"
	"net/http"
)

var sseClients = make(map[chan string]bool)

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string)
	sseClients[messageChan] = true

	defer func() {
		delete(sseClients, messageChan)
		close(messageChan)
	}()

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func BroadcastSSEUpdate(data map[string]int64) {
	message := fmt.Sprintf(`{"dog": %d, "cat": %d}`, data["dog"], data["cat"])
	for client := range sseClients {
		client <- message
	}
}
