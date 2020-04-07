package gtwebsocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"go-jw-toggl/jwtoggl/gtoggl/gtproject"
	"go-jw-toggl/jwtoggl/gtoggl/gttag"
	"go-jw-toggl/jwtoggl/gtoggl/gttimentry"
	"log"
	"net/http"
	"time"
)

const (
	DefaultUrl   = "wss://stream.toggl.com/ws'"
	ActionInsert = "INSERT"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"

	ModelTimeEntry = "time_entry"
	ModelProject   = "project"
	ModelClient    = "client"
	ModelTag       = "tag"
)

var notAuthenticated = errors.New("no session after auth WS message")

func (mp MsgPing) OK() bool {
	return mp.Type == "ping"
}

func defaultOnPing(msg MsgPing) error {
	if !msg.OK() {
		return fmt.Errorf("ping: something wrong: %s", msg.Type)
	}
	log.Println("ping")
	return nil
}

func defaultOnTimeEntryAction(action string, entry gttimeentry.TimeEntry) error {
	log.Printf("TimeEntry action: %s %+v\n", action, entry)
	return nil
}

func defaultOnProjectAction(action string, project gtproject.Project) error {
	log.Printf("Project action: %s %+v\n", action, project)
	return nil
}

func defaultOnClientAction(action string, client gtclient.Client) error {
	log.Printf("Client action: %s %+v\n", action, client)
	return nil
}

func defaultOnTagAction(action string, tag gttag.Tag) error {
	log.Printf("Tag action: %s %+v\n", action, tag)
	return nil
}

func closeWs(ws *websocket.Conn) {
	if err := ws.Close(); err != nil {
		log.Println("Close Websocket error:", err)
	}
}

type TogglWebsocketClient struct {
	url               string
	token             string
	ws                *websocket.Conn
	waiter            *time.Timer
	pingTime          time.Time
	onPing            OnPingCallback
	onTimeEntryAction OnTimeEntryActionCallback
	onProjectAction   OnProjectActionCallback
	onClientAction    OnClientActionCallback
	onTagAction       OnTagActionCallback
}

func NewClient(token string) *TogglWebsocketClient {
	return &TogglWebsocketClient{
		url:               DefaultUrl,
		token:             token,
		onPing:            defaultOnPing,
		onTimeEntryAction: defaultOnTimeEntryAction,
		onProjectAction:   defaultOnProjectAction,
		onClientAction:    defaultOnClientAction,
		onTagAction:       defaultOnTagAction,
	}
}

func (c *TogglWebsocketClient) OnPing(callback OnPingCallback) {
	c.onPing = callback
}

func (c *TogglWebsocketClient) OnTimeEntryAction(callback OnTimeEntryActionCallback) {
	c.onTimeEntryAction = callback
}

func (c *TogglWebsocketClient) OnProjectAction(callback OnProjectActionCallback) {
	c.onProjectAction = callback
}

func (c *TogglWebsocketClient) OnClientAction(callback OnClientActionCallback) {
	c.onClientAction = callback
}

func (c *TogglWebsocketClient) OnTagAction(callback OnTagActionCallback) {
	c.onTagAction = callback
}

func (c *TogglWebsocketClient) Dial() error {
	log.Println("Dialing WS...")
	ws, resp, err := websocket.DefaultDialer.Dial(c.url, http.Header{
		"Origin": {"https://jw-toggl.com/app"},
	})
	if err != nil {
		log.Printf("Failed WS connect:%+v %+v", err, resp)
		return err
	}
	c.ws = ws
	log.Println("WS connected")
	return nil
}

func (c *TogglWebsocketClient) Listen(ctx context.Context) error {
	timeout := time.Second
	for {
		if err := c.Dial(); err != nil {
			time.Sleep(timeout)
			timeout *= 2
			continue
		} else {
			timeout = time.Second
		}
		if err := c.authenticate(); err != nil {
			if err == notAuthenticated {
				log.Println("WS Auth error:", err)
				return err
			} else {
				closeWs(c.ws)
				continue
			}
		}

		disconnected := make(chan error)
		go func() {
			disconnected <- c.listen(ctx)
		}()

		select {
		case <-disconnected:
			continue
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *TogglWebsocketClient) listen(ctx context.Context) error {
	webSocketClosed := make(chan error)
	messages := make(chan []byte)
	go func() {
		defer close(webSocketClosed)
		for {
			if mt, data, err := c.ws.ReadMessage(); err != nil {
				log.Println("Read Websocket error:", err)
				webSocketClosed <- err
				closeWs(c.ws)
				return
			} else {
				if mt != websocket.TextMessage {
					log.Println("Unexpected message type:", mt)
				}
				messages <- data
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			closeWs(c.ws)
			return nil
		case err := <-webSocketClosed:
			log.Println("Websocket closed")
			return err
		case msg := <-messages:
			if err := c.handleMessage(msg); err != nil {
				log.Println("Message handle error:", err)
				closeWs(c.ws)
				return err
			}
		}
	}
}

func (c *TogglWebsocketClient) handleMessage(msg []byte) error {
	var guess map[string]interface{}
	if err := json.Unmarshal(msg, &guess); err != nil {
		return err
	}
	if _, ok := guess["type"]; ok {
		return c.handlePingMessage(msg)
	}
	if _, ok := guess["action"]; ok {
		return c.handleActionMessage(msg)
	}
	return errors.New("unknown message")
}

func (c *TogglWebsocketClient) handlePingMessage(msg []byte) error {
	if c.waiter != nil {
		c.waiter.Stop()
		c.waiter = nil
	}
	now := time.Now()
	if !c.pingTime.IsZero() {
		interval := now.Sub(c.pingTime)
		c.waiter = time.AfterFunc(interval*3, c.handlePingTimeout)
	}
	c.pingTime = now
	var pingMsg MsgPing
	if err := json.Unmarshal(msg, &pingMsg); err != nil {
		return err
	}
	if err := c.onPing(pingMsg); err != nil {
		return err
	} else {
		pingMsg.Type = "pong"
		return c.ws.WriteJSON(pingMsg)
	}
}

func (c *TogglWebsocketClient) authenticate() error {
	type auth struct {
		Type     string `json:"type"`
		ApiToken string `json:"api_token"`
	}

	data := auth{
		Type:     "authenticate",
		ApiToken: c.token,
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = c.ws.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		return err
	}

	var msg struct {
		Session string `json:"session_id"`
	}

	mt, bytes, err := c.ws.ReadMessage()
	if err != nil {
		return err
	}
	if mt != websocket.TextMessage {
		return fmt.Errorf("unexpected auth message type: %d", mt)
	}
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		return err
	}
	log.Printf("Auth: %+v\n", msg)
	if msg.Session == "" {
		return notAuthenticated
	}
	return err
}

func (c *TogglWebsocketClient) handleActionMessage(msg []byte) error {
	var actionMsg MsgAction
	if err := json.Unmarshal(msg, &actionMsg); err != nil {
		return err
	}
	if actionMsg.Model == ModelTimeEntry {
		var timeEntryActionMsg MsgTimeEntryAction
		if err := json.Unmarshal(msg, &timeEntryActionMsg); err != nil {
			return err
		}
		return c.onTimeEntryAction(actionMsg.Action, timeEntryActionMsg.Data)
	}
	if actionMsg.Model == ModelProject {
		var projectActionMsg MsgProjectAction
		if err := json.Unmarshal(msg, &projectActionMsg); err != nil {
			return err
		}
		return c.onProjectAction(actionMsg.Action, projectActionMsg.Data)

	}
	if actionMsg.Model == ModelClient {
		var clientActionMsg MsgClientAction
		if err := json.Unmarshal(msg, &clientActionMsg); err != nil {
			return err
		}
		return c.onClientAction(actionMsg.Action, clientActionMsg.Data)

	}
	if actionMsg.Model == ModelTag {
		var tagActionMsg MsgTagAction
		if err := json.Unmarshal(msg, &tagActionMsg); err != nil {
			return err
		}
		return c.onTagAction(actionMsg.Action, tagActionMsg.Data)

	}
	return fmt.Errorf("unknown model action: %s", actionMsg.Model)
}

func (c *TogglWebsocketClient) handlePingTimeout() {
	log.Println("WebSocket ping timeout")
	closeWs(c.ws)
	c.waiter = nil
}
