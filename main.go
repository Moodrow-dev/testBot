package main

import (
	"database/sql"
	"log"
)

const BOT_TOKEN = "7904777920:AAHYe1_LxZmpYp6M5k5Xlll5P_1uPv34gQo"

func main() {
	db, err := sql.Open("sqlite3", "./bot.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// НЕ ТРОГАТЬ ВСЕ ЧТО ВЫШЕ

	_, _, bh, err := createBotAndPoll()

	chatInit(&bh, db)
	changeNumDenum(&bh, db)
	changeWeek(&bh, db)
	changeTitle(&bh, db)

	_ = bh.Start()
}
