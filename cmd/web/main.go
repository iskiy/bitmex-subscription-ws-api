package main

import (
	"bitmex-subscription-ws-api/pkg/config"
	"bitmex-subscription-ws-api/pkg/gate"
	"bitmex-subscription-ws-api/pkg/handler"
	"github.com/gin-gonic/gin"
	"log"
)


func main(){
	cfg, err := config.NewConfig(".env")
	if err != nil{
		log.Fatalln(err.Error())
	}

	gate, err:= gate.NewWebSocketGateway(cfg.ApiKey, cfg.ApiSecret)
	if err != nil{
		log.Fatalln(err)
	}

	manager := handler.NewWSAPIManager(gate)
	go manager.Run()

	r := gin.Default()
	r.GET("/ws", manager.WsEndpoint)
	r.Run()
}