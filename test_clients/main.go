package main

import (
	"bitmex-subscription-ws-api/pkg/handler"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

func main(){
	testClient(1, []string{"SOLUSDT", "FTMUSDT", "XBTUSDT", "ETHUSDT"})
	testClient(2, []string{"DOGEUSDT", "BCHUSDT", "ADAUSDT", "XBTUSDT"})

	fmt.Scanln()
}

func testClient(idx int, symbols []string){
	client, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil{
		log.Fatalln(err)
	}

	go func() {
		req := handler.WsRequest{Action: "subscribe", Symbols: symbols}
		err = client.WriteJSON(req)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
		for {
			_, message, err := client.ReadMessage()
			if err != nil{
				log.Println(err.Error())
				continue
			}
			log.Printf("Client %d received: %s\n", idx, string(message))
		}
	}()
}
