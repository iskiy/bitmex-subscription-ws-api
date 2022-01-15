package handler

import (
	"bitmex-subscription-ws-api/pkg/gate"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

type WsRequest struct {
	Action 	string 		`json:"action"`
	Symbols []string 	`json:"symbols,omitempty"`
}

type BitMexResponse struct {
	Data []WsResponse `json:"Data"`
}

type WsResponse struct {
	Timestamp string  `json:"timestamp"`
	Symbol	  string  `json:"symbol"`
	Price	  float64 `json:"lastPrice"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSAPIManager struct {
	gate          *gate.WebSocketGate
	Subscriptions map[string]map[*websocket.Conn]struct{}
	mutex         sync.RWMutex
}

func NewWSAPIManager(wsGateway *gate.WebSocketGate) *WSAPIManager {
	subscriptions := make(map[string]map[*websocket.Conn]struct{})
	return &WSAPIManager{gate: wsGateway, Subscriptions: subscriptions}
}

func (m *WSAPIManager) WsEndpoint(c *gin.Context){
	ws, err := upgradeConnection.Upgrade(c.Writer, c.Request, nil)
	if err != nil{
		log.Println(err)
		return
	}
	log.Println("Client connected to endpoint")
	go m.handleWSRequest(ws)
}

func (m *WSAPIManager) handleWSRequest(conn  *websocket.Conn){
	for {
		var request WsRequest
		err := conn.ReadJSON(&request)
		if err != nil {
			log.Println(err)
			return
		}
		if request.Action == "subscribe" {
			m.handleSubscribeAction(&request, conn)
		} else if request.Action == "unsubscribe" {
			m.handleUnsubscribeAction(conn)
		}
	}
}

func (m *WSAPIManager) handleSubscribeAction(request *WsRequest, conn *websocket.Conn){
	m.mutex.Lock()
	for _, s := range request.Symbols{
		if(m.Subscriptions[s]) == nil{
			m.Subscriptions[s] = make(map[*websocket.Conn]struct{})
			m.gate.SendSubCommand("subscribe", []string{s})
		}
		m.Subscriptions[s][conn] = struct{}{}
	}
	m.mutex.Unlock()
	log.Println("Subscribed")
}

func (m *WSAPIManager) handleUnsubscribeAction(conn *websocket.Conn){
	m.mutex.Lock()
	for s, v := range m.Subscriptions{
		delete(v, conn)
		if len(s) == 0{
			m.gate.SendSubCommand("unsubscribe", []string{s})
		}
	}
	m.mutex.Unlock()
}

func (m *WSAPIManager) Run() {
	for {
		response, err := m.gate.ReadUpdate()
		if err != nil{
			log.Println(err)
			continue
		}
		go m.handleResponse(response)
	}
}

func (m *WSAPIManager) handleResponse(response string){
	if !isLastPriceUpdate(&response){
		return
	}
	log.Println("Response: " + response)
	var data BitMexResponse
	err := json.Unmarshal([]byte(response), &data)
	if err != nil {
		log.Println(err)
	}

	if len(data.Data) < 1{
		log.Println(response)
		return
	}
	apiResponse := data.Data[0]
	m.mutex.RLock()
	for k := range m.Subscriptions[apiResponse.Symbol]{

		if err = k.WriteJSON(apiResponse); err != nil{
			log.Println(err)
		} else{
			log.Printf("Sended response: %s\n", apiResponse)
		}
	}
	m.mutex.RUnlock()
}

func isLastPriceUpdate(message *string) bool{
	return strings.Contains(*message, "lastPrice\":") && !strings.Contains(*message, "\"partial\"")
}