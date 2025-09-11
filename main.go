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
)

func main() {
	db, err := sql.Open("sqlite", "./bot.db")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// НЕ ТРОГАТЬ ВСЕ ЧТО ВЫШЕ

	bot, bh, err := CreateBotAndPoll()

	c := cron.New()

	ChangeAllWeeks := func() {
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

	TolstobrowConnection := func() {
		ids, err1 := pickOverIds(db)
		if err1 != nil {
			log.Println(err1)
		}
		for _, id := range ids {
			println(id)
			chat := read(id, db)
			if chat.UseTolstobrow {
				photo, _ := os.ReadFile("connection.jpg")
				_, err2 := bot.SendPhoto(context.Background(), &telego.SendPhotoParams{MessageThreadID: chat.InfoThread, ParseMode: telego.ModeMarkdownV2, ChatID: telego.ChatID{ID: chat.ID}, Photo: tu.FileFromReader(bytes.NewReader(photo), "connection"), Caption: "[Tolstobrow connection](https://edu.vsu.ru/mod/bigbluebuttonbn/view.php?id=1095331)"})
				if err2 != nil {
					//fmt.Print("ОШИБКА")
					log.Println(err2)
				}
			}
		}
	}

	c.AddFunc("0 0 0 * * 1", ChangeAllWeeks)
	c.AddFunc("0 30 18 * * 3", TolstobrowConnection)

	adminCmds := []telego.BotCommand{
		{Command: "init", Description: "Проинициализировать бота"},
		{Command: "changeweek", Description: "Вручную сменить ЧИСЛ/ЗНАМ"},
		{Command: "changeweektitle", Description: "Сменить названия ЧИСЛ/ЗНАМ(использовать без [])"},
		{Command: "changetitle", Description: "Сменить название чата(без ЧИСЛ/ЗНАМ)"},
		{Command: "setusers", Description: "Установить список пользователей(для пинга)"},
		{Command: "ping", Description: "Пинг всех(установленных) юзеров(через @)"},
		{Command: "setmainthread", Description: "Установить чат(только для суперчатов) для уведомлений(напр. Толстобров)"},
		{Command: "tolstobrow", Description: "Включить/выключить оповещения на пары Толстоброва"},
	}

	userCmds := []telego.BotCommand{
		{Command: "ping", Description: "Пинг всех(установленных) юзеров"},
	}

	//bot.DeleteMyCommands(context.Background(), nil)
	err = bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{Commands: adminCmds, Scope: telego.BotCommandScope(&telego.BotCommandScopeAllChatAdministrators{"all_chat_administrators"})})
	if err != nil {
		log.Println(err)
	}
	err = bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{Commands: userCmds, Scope: &telego.BotCommandScopeAllGroupChats{"all_group_chats"}})
	if err != nil {
		log.Println(err)
	}

	ChatInit(bh, db)
	ChangeNumDenum(bh, db)
	ChangeWeek(bh, db)
	ChangeTitle(bh, db)
	SetUsers(bh, db)
	SetMainThread(bh, db)
	Ping(bh, db)
	Tolstobrow(bh, db)

	// Не трогать
	AddNewPeople(bh, db)
	DelLeftPeople(bh, db)
	// \/\/\/\/\/\/\/\/\/\
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
