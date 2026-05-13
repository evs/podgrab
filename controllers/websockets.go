package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type EnqueuePayload struct {
	ItemIds   []string `json:"itemIds"`
	PodcastId string   `json:"podcastId"`
	TagIds    []string `json:"tagIds"`
}

var activePlayers = make(map[*websocket.Conn]string)
var allConnections = make(map[*websocket.Conn]string)

var broadcast = make(chan Message) // broadcast channel

var playersMutex sync.RWMutex
var connectionsMutex sync.RWMutex

type Message struct {
	Identifier  string          `json:"identifier"`
	MessageType string          `json:"messageType"`
	Payload     string          `json:"payload"`
	Connection  *websocket.Conn `json:"-"`
}

func Wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := context.Background()
	for {
		var mess Message
		err := wsjson.Read(ctx, conn, &mess)
		if err != nil {
			isPlayer := false
			playersMutex.RLock()
			isPlayer = activePlayers[conn] != ""
			playersMutex.RUnlock()
			if isPlayer {
				playersMutex.Lock()
				delete(activePlayers, conn)
				playersMutex.Unlock()
				broadcast <- Message{
					MessageType: "PlayerRemoved",
					Identifier:  mess.Identifier,
				}
			}
			connectionsMutex.Lock()
			delete(allConnections, conn)
			connectionsMutex.Unlock()
			break
		}
		mess.Connection = conn
		connectionsMutex.Lock()
		allConnections[conn] = mess.Identifier
		connectionsMutex.Unlock()
		broadcast <- mess
	}
}

func HandleWebsocketMessages() {
	ctx := context.Background()
	for {
		msg := <-broadcast

		switch msg.MessageType {
		case "RegisterPlayer":
			playersMutex.Lock()
			activePlayers[msg.Connection] = msg.Identifier
			playersMutex.Unlock()
			connectionsMutex.RLock()
			for connection := range allConnections {
				wsjson.Write(ctx, connection, Message{
					Identifier:  msg.Identifier,
					MessageType: "PlayerExists",
				})
			}
			connectionsMutex.RUnlock()
			fmt.Println("Player Registered")
		case "PlayerRemoved":
			connectionsMutex.RLock()
			for connection := range allConnections {
				wsjson.Write(ctx, connection, Message{
					Identifier:  msg.Identifier,
					MessageType: "NoPlayer",
				})
			}
			connectionsMutex.RUnlock()
			fmt.Println("Player Removed")
		case "Enqueue":
			var payload EnqueuePayload
			fmt.Println(msg.Payload)
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err == nil {
				items := getItemsToPlay(payload.ItemIds, payload.PodcastId, payload.TagIds)
				var player *websocket.Conn
				playersMutex.RLock()
				for connection, id := range activePlayers {
					if msg.Identifier == id {
						player = connection
						break
					}
				}
				playersMutex.RUnlock()
				if player != nil {
					payloadStr, err := json.Marshal(items)
					if err == nil {
						wsjson.Write(ctx, player, Message{
							Identifier:  msg.Identifier,
							MessageType: "Enqueue",
							Payload:     string(payloadStr),
						})
					}
				}
			} else {
				fmt.Println(err.Error())
			}
		case "Register":
			var player *websocket.Conn
			playersMutex.RLock()
			for connection, id := range activePlayers {
				if msg.Identifier == id {
					player = connection
					break
				}
			}
			playersMutex.RUnlock()

			if player == nil {
				fmt.Println("Player Not Exists")
				wsjson.Write(ctx, msg.Connection, Message{
					Identifier:  msg.Identifier,
					MessageType: "NoPlayer",
				})
			} else {
				wsjson.Write(ctx, msg.Connection, Message{
					Identifier:  msg.Identifier,
					MessageType: "PlayerExists",
				})
			}
		}
	}
}
