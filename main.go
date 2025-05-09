package main

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/robfig/cron"
	"log"
	"os"
	"time"
)

func main() {
	db, err := sql.Open("sqlite3", "./bot.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// НЕ ТРОГАТЬ ВСЕ ЧТО ВЫШЕ

	bot, bh, err := createBotAndPoll()

	loc, _ := time.LoadLocation("Europe/Moscow")
	c := cron.NewWithLocation(loc)

	changeAllWeeks := func() {
		ids, err1 := pickOverIds(db)
		if err1 != nil {
			log.Println(err1)
		}
		for _, id := range ids {
			println(id)
			err2 := changeWeekMain(telego.ChatID{ID: id}, bot, db)
			if err2 != nil {
				//fmt.Print("ОШИБКА")
				log.Println(err2)
			}
		}
	}

	photo, _ := os.ReadFile("connection.jpg")

	tolstobrowConnection := func() {
		ids, err1 := pickOverIds(db)
		if err1 != nil {
			log.Println(err1)
		}
		for _, id := range ids {
			println(id)
			chatID := telego.ChatID{ID: id}
			_, err2 := bot.SendPhoto(context.Background(), &telego.SendPhotoParams{ParseMode: telego.ModeMarkdownV2, ChatID: chatID, Photo: tu.FileFromReader(bytes.NewReader(photo), "connection"), Caption: "[Tolstobrow connection](https://edu.vsu.ru/mod/bigbluebuttonbn/view.php?id=1095331)"})
			if err2 != nil {
				//fmt.Print("ОШИБКА")
				log.Println(err2)
			}
		}
	}

	c.AddFunc("0 0 0 * * 1", changeAllWeeks)
	c.AddFunc("0 30 18 * * 3", tolstobrowConnection)

	tolstobrowConnection()

	chatInit(&bh, db)
	changeNumDenum(&bh, db)
	changeWeek(&bh, db)
	changeTitle(&bh, db)
	setUsers(&bh, db)
	ping(&bh, db)
	addNewPeople(&bh, db)
	delLeftPeople(&bh, db)

	go c.Start()
	defer c.Stop()

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
