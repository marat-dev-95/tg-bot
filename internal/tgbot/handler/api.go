package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type User struct {
	Id        int    `db:"id"`
	Firstname string `db:"FirstName"`
	Tg_id     int    `db:"Tg_id"`
}

type sendLogInput struct {
	Tag     string `form:"tag" binding:"required"`
	Message string `form:"message" binding:"required"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func sendLog(c *gin.Context) {
	var input sendLogInput

	token, _ := os.LookupEnv("TOKEN")

	bot, _ := tgbotapi.NewBotAPI(token)

	if err := c.ShouldBindQuery(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		fmt.Println(err)
	} else {
		fmt.Println(input)
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

}
