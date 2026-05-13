package controllers

import (
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestWebSocketDataRace(t *testing.T) {
	// Reset maps for deterministic test
	playersMutex.Lock()
	activePlayers = make(map[*websocket.Conn]string)
	playersMutex.Unlock()

	connectionsMutex.Lock()
	allConnections = make(map[*websocket.Conn]string)
	connectionsMutex.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := &websocket.Conn{} // mock — we only test map ops, not real WS
			id := "player-" + string(rune('A'+i%26))
			playersMutex.Lock()
			activePlayers[conn] = id
			playersMutex.Unlock()

			connectionsMutex.Lock()
			allConnections[conn] = id
			connectionsMutex.Unlock()

			time.Sleep(1 * time.Millisecond)

			playersMutex.RLock()
			_ = activePlayers[conn]
			playersMutex.RUnlock()

			connectionsMutex.RLock()
			_ = allConnections[conn]
			connectionsMutex.RUnlock()

			playersMutex.Lock()
			delete(activePlayers, conn)
			playersMutex.Unlock()

			connectionsMutex.Lock()
			delete(allConnections, conn)
			connectionsMutex.Unlock()
		}(i)
	}
	wg.Wait()
}
