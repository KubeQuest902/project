package projectserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mediocregopher/radix/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var mutex = &sync.Mutex{}

const broadcastChannel = "websocket_broadcast"

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	sendCurrentCounts(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
	}
}

func sendCurrentCounts(conn *websocket.Conn) {
	var dogCount, catCount string

	err := RedisPool.Do(radix.Cmd(&dogCount, "GET", "dog_count"))
	if err != nil || dogCount == "" {
		dogCount = "0"
	}

	err = RedisPool.Do(radix.Cmd(&catCount, "GET", "cat_count"))
	if err != nil || catCount == "" {
		catCount = "0"
	}

	var dogCountInt, catCountInt int64
	if dogCount != "" {
		if val, err := strconv.ParseInt(dogCount, 10, 64); err == nil {
			dogCountInt = val
		}
	}
	if catCount != "" {
		if val, err := strconv.ParseInt(catCount, 10, 64); err == nil {
			catCountInt = val
		}
	}

	data := map[string]int64{
		"dog": dogCountInt,
		"cat": catCountInt,
	}

	conn.WriteJSON(data)
}

func BroadcastUpdates(data map[string]int64) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling broadcast data: %v", err)
		return
	}

	err = RedisPool.Do(radix.Cmd(nil, "PUBLISH", broadcastChannel, string(jsonData)))
	if err != nil {
		log.Printf("Error publishing to Redis: %v", err)
	}
}

func StartWebSocketBroadcaster() {
	for {
		err := subscribeAndListen()
		if err != nil {
			log.Printf("Redis pub/sub connection error: %v, retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func subscribeAndListen() error {
	conn, err := radix.Dial("tcp", RedisHost+":"+RedisPort, radix.DialAuthPass(RedisPassword))
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("WebSocket broadcaster connected to Redis pub/sub channel: %s", broadcastChannel)

	if err := conn.Do(radix.Cmd(nil, "SUBSCRIBE", broadcastChannel)); err != nil {
		return err
	}

	for {
		var resp []interface{}
		if err := conn.Do(radix.Cmd(&resp, "BLPOP", "dummy", "0")); err != nil {
			var message []string
			if err := conn.Do(radix.Cmd(&message, "BLPOP", "dummy", "1")); err != nil {
				var rawResp interface{}
				if err := conn.Do(radix.Cmd(&rawResp, "PING")); err != nil {
					return err
				}

				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		break
	}

	return nil
}

func StartWebSocketBroadcasterPolling() {
	var lastDogCount, lastCatCount int64 = -1, -1

	log.Println("Starting WebSocket broadcaster with polling...")

	for {
		var dogCountStr, catCountStr string
		var dogCount, catCount int64

		err := RedisPool.Do(radix.Cmd(&dogCountStr, "GET", "dog_count"))
		if err != nil || dogCountStr == "" {
			dogCount = 0
		} else {
			if val, err := strconv.ParseInt(dogCountStr, 10, 64); err == nil {
				dogCount = val
			}
		}

		err = RedisPool.Do(radix.Cmd(&catCountStr, "GET", "cat_count"))
		if err != nil || catCountStr == "" {
			catCount = 0
		} else {
			if val, err := strconv.ParseInt(catCountStr, 10, 64); err == nil {
				catCount = val
			}
		}

		if dogCount != lastDogCount || catCount != lastCatCount {
			data := map[string]int64{
				"dog": dogCount,
				"cat": catCount,
			}

			log.Printf("Broadcasting update: %+v", data)

			mutex.Lock()
			clientCount := len(clients)
			for client := range clients {
				err := client.WriteJSON(data)
				if err != nil {
					client.Close()
					delete(clients, client)
				}
			}
			mutex.Unlock()

			log.Printf("Broadcasted to %d clients", clientCount)

			lastDogCount = dogCount
			lastCatCount = catCount
		}

		time.Sleep(200 * time.Millisecond)
	}
}
