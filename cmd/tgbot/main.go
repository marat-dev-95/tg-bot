package main

import (
	"github.com/marat-dev-95/tg-bot/internal/tgbot/bot"
	"github.com/marat-dev-95/tg-bot/internal/tgbot/handler"
	"github.com/marat-dev-95/tg-bot/internal/tgbot/server"
)

func main() {
	server := new(server.Server)
	handler := new(handler.Handler)
	go bot.Run()
	server.Run("80", handler.InitRoutes())

}
