package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
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

const TOKEN = "2095241662:AAHdJc9IfMBFMi7aS07Ds2XU8L14FFBHWgE"

func Run() {
	bot, err := tgbotapi.NewBotAPI(TOKEN)

	if err != nil {
		log.Panic(err)
	}

	db, err := sqlx.Connect("sqlite3", "base.db")

	if err != nil {
		log.Fatalln(err)
	}
	//db.MustExec("DROP TABLE tags")
	//db.MustExec("CREATE TABLE tg_users(id INTEGER PRIMARY KEY AUTOINCREMENT, FirstName VARCHAR(255), Tg_id INTEGER)")
	//db.MustExec("CREATE TABLE tags(id INTEGER PRIMARY KEY AUTOINCREMENT, User_id INTEGER, Tag VARCHAR(255))")

	bot.Debug = true

	log.Printf("Authorized on account %s", &bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		users := []User{
			{Firstname: update.Message.From.UserName, Tg_id: update.Message.From.ID},
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		var CurrUser = User{}
		err = db.Get(&CurrUser, "SELECT * FROM tg_users WHERE Tg_id=$1", update.Message.From.ID)

		if update.Message.IsCommand() {

			switch update.Message.Command() {
			case "start":
				if err != nil {
					db.NamedExec("INSERT INTO tg_users(FirstName, Tg_id) VALUES(:firstname, :tg_id)", users)
				}
				msg.Text = "Выполните команду /add_tag #tag чтобы подписаться на тег "
			case "add_tag":
				if err != nil {
					msg.Text = "Выполните сначало /start"
				} else {
					tag := Tag{}

					err = db.Get(&tag, "SELECT * FROM tags WHERE  Tag=$1 AND User_id=$2", update.Message.CommandArguments(), CurrUser.Id)
					fmt.Printf("%#v\n", tag)
					if err != nil {
						db.MustExec("INSERT INTO tags(Tag, User_id) VALUES($1, $2)", update.Message.CommandArguments(), CurrUser.Id)
						msg.Text = "Вы успешно подписаны на тег " + update.Message.CommandArguments()
					} else {
						msg.Text = "Вы уже подписаны на тег " + update.Message.CommandArguments()
					}
				}
			case "delete_tag":
				if err != nil {
					msg.Text = "Вы не подписаны на этот тег"
				} else {
					db.MustExec("DELETE FROM tags WHERE User_id=$1 AND tag=$1", CurrUser.Id, update.Message.CommandArguments())
					msg.Text = "Вы отписались от тега " + update.Message.CommandArguments()
				}
			case "stop":
				if err != nil {
					msg.Text = "Вы не подписаны на теги"
				} else {
					db.MustExec("DELETE FROM tags WHERE User_id=$1", CurrUser.Id)
					msg.Text = "Вы успешно отписались"
				}
			}

			bot.Send(msg)
		}
	}
}
