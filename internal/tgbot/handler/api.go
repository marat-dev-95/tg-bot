package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Id        int    `db:"id"`
	Firstname string `db:"FirstName"`
	Tg_id     int    `db:"Tg_id"`
}

type sendLogInput struct {
	Tag     string `json:"tag" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	//c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

const TOKEN = "2095241662:AAHdJc9IfMBFMi7aS07Ds2XU8L14FFBHWgE"

func sendLog(c *gin.Context) {
	var input sendLogInput

	bot, err := tgbotapi.NewBotAPI(TOKEN)

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
	}

	db, err := sqlx.Connect("sqlite3", "base.db")

	if err != nil {
		log.Fatalln(err)
	}

	var users []User

	err = db.Select(&users, "SELECT tg_users.id, tg_users.FirstName, tg_users.Tg_id FROM tg_users INNER JOIN tags ON tags.User_id=tg_users.id AND tags.Tag=?", input.Tag)

	for _, user := range users {
		msg := tgbotapi.NewMessage(int64(user.Tg_id), input.Message)
		bot.Send(msg)
	}

	c.IndentedJSON(http.StatusOK, map[string]interface{}{
		"message": "success",
	})
}
