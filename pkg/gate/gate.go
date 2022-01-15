package gate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	BitmexURL = "wss://ws.testnet.bitmex.com"
	Verb      = "GET"
	Endpoint  = "/realtime"
)

var (
	expires = time.Now().Add(3* time.Minute)
)

type WebSocketGate struct {
	wsConn *websocket.Conn
}

func NewWebSocketGateway(apiKey, apiSecret string) (*WebSocketGate, error) {
	signature, err := generateSignature(apiSecret, Verb, Endpoint, expires.Unix())
	if err != nil{
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(BitmexURL+Endpoint, nil)
	if err != nil {
		return nil, err
	}

	if _, err = ReadResponse(conn); err != nil{
		return nil, err
	}

	auth := BitmexCommand{Op: "authKeyExpires",
		Args: []interface{}{apiKey, expires.Unix(), signature}}

	if err := conn.WriteJSON(auth); err != nil {
		return nil, err
	}

	if _, err = ReadResponse(conn); err != nil{
		return nil, err
	}

	return &WebSocketGate{wsConn: conn}, nil
}

// SendSubCommand used for send subscribe/unsubscribe command to BitMex
func (g *WebSocketGate) SendSubCommand(command string,symbols []string){
	args := make([]interface{}, len(symbols))
	for i, _ := range args{
		args[i] = "instrument:" + symbols[i]
	}
	g.sendCommand(command, args)

}

func (g *WebSocketGate) sendCommand(op string, args []interface{}){
	command := BitmexCommand{op, args}
	err := g.wsConn.WriteJSON(command)
	if err != nil {
		log.Println(err)
		return
	}
}

func (g *WebSocketGate) ReadUpdate() (string, error){
	return ReadResponse(g.wsConn)
}

func ReadResponse(conn *websocket.Conn) (string,error){
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "",err
	}
	return string(message), nil
}

func generateSignature(secret, verb, url string, expires int64) (string, error) {
	message := fmt.Sprintf("%s%s%d", verb, url, expires)
	sig := hmac.New(sha256.New, []byte(secret))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}

	return hex.EncodeToString(sig.Sum(nil)), nil
}
