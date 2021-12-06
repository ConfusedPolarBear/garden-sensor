package websocket

import (
	"container/list"
	"errors"
	"net"
	"net/http"
	"regexp"
	"sync"
	"syscall"

	"github.com/ConfusedPolarBear/garden/internal/db"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebSocketMessage struct {
	Type string
	Data interface{}
}

// List of all connected websockets. A slice is not used here to allow for easy removal of stale sockets.
var wsClients list.List

// Concurrent writes to websockets are illegal
var websocketLock sync.Mutex

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Don't check the port when validating the origin to allow for development setups to function
			// TODO: what happens if a header doesn't have a port in it?

			rawHost := r.Host
			rawOrigin := r.Header.Get("Origin")

			logrus.Debugf("[server] raw host is %s and raw origin is %s", rawHost, rawOrigin)

			// Split the requested host from the port
			host, _, err := net.SplitHostPort(rawHost)
			if err != nil {
				logrus.Warnf("[server] unable to split host and port, blocking websocket")
				return false
			}

			// Origins have the HTTP(S) protocol prepended to them so remove it before attempting to split it
			rawOrigin = regexp.MustCompile("^https?://").ReplaceAllString(rawOrigin, "")

			origin, _, err := net.SplitHostPort(rawOrigin)
			if err != nil {
				logrus.Warnf("[server] unable to split origin and port, blocking websocket")
				return false
			}

			okay := host == origin
			logrus.Debugf("[server] host is %s, origin is %s, websocket ok: %t", host, origin, okay)

			return okay
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Warnf("[server] unable to upgrade connection to websocket: %s", err)
		return
	}

	// Send all systems for the first update
	conn.WriteJSON(WebSocketMessage{
		Type: "register",
		Data: db.GetAllSystems(),
	})

	wsClients.PushBack(conn)
}

func BroadcastWebsocketMessage(messageType string, data interface{}) {
	websocketLock.Lock()
	defer websocketLock.Unlock()

	msg := WebSocketMessage{
		Type: messageType,
		Data: data,
	}

	var staleConnections []*list.Element

	// Send this message to all registered websockets
	for e := wsClients.Front(); e != nil; e = e.Next() {
		if e.Value == nil {
			wsClients.Remove(e)
			logrus.Warn("[server] removed nil websocket")
			continue
		}

		c := e.Value.(*websocket.Conn)

		if err := c.WriteJSON(msg); err != nil {
			// A broken pipe error occurs when the browser tab/window is closed. Since it is an expected error, don't log it.
			if !errors.Is(err, syscall.EPIPE) {
				logrus.Warnf("[server] unable to send websocket message to %s: %s", c.LocalAddr(), err)
			}

			c.Close()

			// Push the element into a slice for later cleanup
			staleConnections = append(staleConnections, e)
		}
	}

	// Remove all broken websockets from the list
	for _, c := range staleConnections {
		wsClients.Remove(c)
	}
}
