package main

import (
	"database/sql"
	"log"
)

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
	setUsers(&bh, db)
	ping(&bh, db)
	addNewPeople(&bh, db)
	delLeftPeople(&bh, db)

	_ = bh.Start()
}
