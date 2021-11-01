package bot

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id             int    `db:"id"`
	Firstname      string `db:"FirstName"`
	Tg_id          int    `db:"Tg_id"`
	Auto_subscribe int    `db:"auto_subscribe"`
}

type Tag struct {
	Id      int    `db:"id"`
	Tag     string `db:"Tag"`
	User_id int    `db:"User_id"`
}

func Run() {
	token, _ := os.LookupEnv("TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Panic(err)
	}

	db, err := sqlx.Connect("sqlite3", "base.db")

	if err != nil {
		log.Fatalln(err)
	}
	//db.MustExec("ALTER TABLE tg_messages ADD COLUMN tag VARCHAR(255)")
	//db.MustExec("ALTER TABLE tg_messages DROP COLUMN tag_id")
	//db.MustExec("ALTER TABLE tg_users ADD auto_subscribe INTEGER NOT NULL DEFAULT(0)")
	//db.MustExec("CREATE TABLE tg_messages(id INTEGER PRIMARY KEY AUTOINCREMENT, tag_id INTEGER, message TEXT)")
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
					_, err = db.NamedExec("INSERT INTO tg_users(FirstName, Tg_id) VALUES(:FirstName, :Tg_id)", users)
					fmt.Println(err)
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

						var users []User

						err = db.Select(&users, "SELECT * FROM tg_users WHERE auto_subscribe=1")
						for _, user := range users {
							if user.Id == CurrUser.Id {
								continue
							}
							db.MustExec("INSERT INTO tags(User_id, Tag) VALUES($1,$2)", user.Id, update.Message.CommandArguments())
						}
					} else {
						msg.Text = "Вы уже подписаны на тег " + update.Message.CommandArguments()
					}
				}
			case "delete_tag":
				if err != nil {
					msg.Text = "Вы не подписаны на этот тег"
				} else {
					db.MustExec("DELETE FROM tags WHERE User_id=$1 AND tag=$2", CurrUser.Id, update.Message.CommandArguments())
					msg.Text = "Вы отписались от тега " + update.Message.CommandArguments()
				}
			case "stop":
				if err != nil {
					msg.Text = "Вы не подписаны на теги"
				} else {
					db.MustExec("DELETE FROM tags WHERE User_id=$1", CurrUser.Id)
					msg.Text = "Вы успешно отписались"
				}
			case "my_tags":
				if err != nil {
					msg.Text = "Вы не подписаны на теги"
				} else {
					var mytags []Tag
					var tag Tag

					mytagsstr := ""

					err = db.Select(&mytags, "SELECT * FROM tags WHERE User_id=$1", CurrUser.Id)
					if len(mytags) == 0 {
						msg.Text = "Вы не подписаны на теги"
					} else {
						for _, tag = range mytags {
							if len(mytagsstr) != 0 {
								mytagsstr += ", "
							}
							mytagsstr = mytagsstr + tag.Tag
						}
						msg.Text = mytagsstr
					}
				}
			case "all_tags":
				var tags []Tag
				tagsstr := ""

				err = db.Select(&tags, "SELECT tags.id,tags.Tag,tags.User_id FROM tags INNER JOIN tg_messages ON tg_messages.tag=tags.Tag GROUP BY tg_messages.tag HAVING COUNT(tg_messages.id) > 0")

				for _, tag := range tags {
					if len(tagsstr) != 0 {
						tagsstr += ", "
					}
					tagsstr += tag.Tag
				}
				msg.Text = tagsstr
			case "add_all_tags":
				if err != nil {
					msg.Text = "Сначало выполните тег /start"
					return
				}
				var allTags []Tag

				err = db.Select(&allTags, "SELECT tags.id,tags.Tag,tags.User_id FROM tags INNER JOIN tg_messages ON tg_messages.tag=tags.Tag GROUP BY tg_messages.tag HAVING COUNT(tg_messages.id) > 0")

				for _, tag := range allTags {
					err = db.Get(&tag, "SELECT * FROM tags WHERE  Tag=$1 AND User_id=$2", tag.Tag, CurrUser.Id)
					fmt.Printf("%#v\n", tag)
					if err != nil {
						db.MustExec("INSERT INTO tags(Tag, User_id) VALUES($1, $2)", tag.Tag, CurrUser.Id)
					}
				}
				msg.Text = "Вы успешно подписаны на все теги"
			case "auto_subscribe":
				if err != nil {
					msg.Text = "Сначало выполните тег /start"
					return
				}
				var command = update.Message.CommandArguments()

				if command == "on" {
					var allTags []Tag

					err = db.Select(&allTags, "SELECT tags.id,tags.Tag,tags.User_id FROM tags INNER JOIN tg_messages ON tg_messages.tag=tags.Tag GROUP BY tg_messages.tag HAVING COUNT(tg_messages.id) > 0")

					for _, tag := range allTags {
						err = db.Get(&tag, "SELECT * FROM tags WHERE  Tag=$1 AND User_id=$2", tag.Tag, CurrUser.Id)
						fmt.Printf("%#v\n", tag)
						if err != nil {
							db.MustExec("INSERT INTO tags(Tag, User_id) VALUES($1, $2)", tag.Tag, CurrUser.Id)
						}
					}
					db.MustExec("UPDATE tg_users SET auto_subscribe=1 WHERE id=$1", CurrUser.Id)
					msg.Text = "Автоподписка включена"
				} else if command == "off" {
					db.MustExec("UPDATE tg_users SET auto_subscribe=0 WHERE id=$1", CurrUser.Id)
					msg.Text = "Автоподписка отключена"
				} else {
					msg.Text = "Неизвестная команда"
				}
			}

			bot.Send(msg)
		}
	}
}
