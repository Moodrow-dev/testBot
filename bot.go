package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"strings"
)

func createBotAndPoll() (telego.Bot, <-chan telego.Update, th.BotHandler, error) {
	bot, err := telego.NewBot(BOT_TOKEN, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
		return telego.Bot{}, nil, th.BotHandler{}, err
	}
	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	return *bot, upd, *bh, nil
}

func chatInit(bh *th.BotHandler, db *sql.DB) Chat {
	var chat Chat
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		//isForum := update.Message.Chat.IsForum
		var err error
		chat, err = read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: err.Error()})
			chat = createChat(int(chatID.ID), update.Message.Chat.Title)
			write(chat, db)
		}
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Чат успешно инициализирован"})
		return nil
	}, th.CommandEqual("init"))
	return chat
}

func changeNumDenum(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		args := strings.Split(update.Message.Text, " ")
		num := args[1]
		denum := args[2]
		chat.Num = num
		chat.Den = denum
		write(chat, db)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf("Успешно.\nЧислитель теперь: %v,\nзнаменатель теперь: %v", num, denum)})
		return nil
	}, th.CommandEqualArgc("changeWeekTitle", 2))
}

func loadPeople(bh *th.BotHandler) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		return nil
	}, th.CommandEqual("people"))
}

func changeWeek(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		oldTitle := update.Message.Chat.Title
		numTitle := fmt.Sprintf("[%v] %v", chat.Den, chat.Title)
		denTitle := fmt.Sprintf("[%v] %v", chat.Num, chat.Title)
		if oldTitle != numTitle {
			changeChatTitle(numTitle, chatID, bh, ctx)
		} else if oldTitle == numTitle {
			changeChatTitle(denTitle, chatID, bh, ctx)
		}
		return nil
	}, th.CommandEqual("changeWeek"))
}

func changeTitle(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		title := strings.Split(update.Message.Text, " ")[1]
		chat.Title = title
		title = fmt.Sprintf("[%v] %v", chat.Num, title)
		changeChatTitle(title, chatID, bh, ctx)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Название изменено успешно"})
		write(chat, db)
		return nil
	}, th.CommandEqualArgc("changeTitle", 1))

}

func changeChatTitle(title string, chatID telego.ChatID, bh *th.BotHandler, ctx *th.Context) {
	err := ctx.Bot().SetChatTitle(ctx, &telego.SetChatTitleParams{ChatID: chatID, Title: title})
	if err != nil {
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "У меня нет прав на смену названия данного чата"})
	}
}
