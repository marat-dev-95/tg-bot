package handler

import (
	"fmt"
	"io/ioutil"
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

type Tag struct {
	Id      int    `db:"id"`
	Tag     string `db:"Tag"`
	User_id int    `db:"User_id"`
}

type sendLogInput struct {
	Tag     string `form:"tag"`
	Message string `form:"message" binding:"required"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type TgMessage struct {
	Id      int    `db:"id"`
	Tag     string `db:"tag"`
	Message string `db:"message"`
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
		return
	}

	fmt.Println(input)
	db, err := sqlx.Connect("sqlite3", "base.db")

	if err != nil {
		log.Fatalln(err)
	}

	var users []User

	if len(input.Tag) == 0 {
		err = db.Select(&users, "SELECT tg_users.id, tg_users.FirstName, tg_users.Tg_id FROM tg_users", input.Tag)
	} else {
		err = db.Select(&users, "SELECT tg_users.id, tg_users.FirstName, tg_users.Tg_id FROM tg_users INNER JOIN tags ON tags.User_id=tg_users.id AND tags.Tag=?", input.Tag)
	}

	if len(users) == 0 {
		newErrorResponse(c, http.StatusBadRequest, "tag not found")
		return
	}
	for _, user := range users {
		msg := tgbotapi.NewMessage(int64(user.Tg_id), input.Message)
		bot.Send(msg)
	}

	db.MustExec("INSERT INTO tg_messages(tag, message) VALUES($1,$2)", input.Tag, input.Message)

	c.IndentedJSON(http.StatusOK, map[string]interface{}{
		"message": "success",
	})

}

func sendLogWithFile(c *gin.Context) {
	var input sendLogInput
	token, _ := os.LookupEnv("TOKEN")
	form, _ := c.Request.MultipartReader()
	part, _ := form.NextPart()

	bot, _ := tgbotapi.NewBotAPI(token)
	fileBytes, _ := ioutil.ReadAll(part)
	fileName := part.FileName()
	documentFileBytes := tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: fileBytes,
	}

	if err := c.ShouldBindQuery(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input")
		fmt.Println(err)
		return
	}

	fmt.Println(input)
	db, err := sqlx.Connect("sqlite3", "base.db")

	if err != nil {
		log.Fatalln(err)
	}
	var users []User
	if len(input.Tag) == 0 {
		err = db.Select(&users, "SELECT tg_users.id, tg_users.FirstName, tg_users.Tg_id FROM tg_users", input.Tag)
	} else {
		err = db.Select(&users, "SELECT tg_users.id, tg_users.FirstName, tg_users.Tg_id FROM tg_users INNER JOIN tags ON tags.User_id=tg_users.id AND tags.Tag=?", input.Tag)
	}

	if len(users) == 0 {
		newErrorResponse(c, http.StatusBadRequest, "tag not found")
		return
	}
	for _, user := range users {
		msg := tgbotapi.NewMessage(int64(user.Tg_id), input.Message)
		bot.Send(msg)
		bot.Send(tgbotapi.NewDocumentUpload(int64(user.Tg_id), documentFileBytes))
	}

	db.MustExec("INSERT INTO tg_messages(tag, message) VALUES($1,$2)", input.Tag, input.Message)

	c.IndentedJSON(http.StatusOK, map[string]interface{}{
		"message": "success",
	})
}
