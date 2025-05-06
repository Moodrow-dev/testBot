package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"slices"
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
		if !fromChat(chatID) {
			return nil
		}
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Положи мамину карточку на место"})
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
		if !fromChat(chatID) {
			return nil
		}
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

func setUsers(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !fromChat(chatID) {
			return nil
		}
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		people := strings.Split(update.Message.Text, " ")[1:]
		flag := true
		for _, peopleStr := range people {
			if peopleStr[0] != '@' {
				flag = false
			}
		}
		if flag {
			chat.Users = []string{}
			chat.Users = append(chat.Users, people...)
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf("Список пользователей\n%v", chat.Users)})
			write(chat, db)
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Неверный формат команды!"})
		}
		return nil
	}, th.CommandEqual("setUsers"))
}

func changeWeek(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !fromChat(chatID) {
			return nil
		}
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
		if !fromChat(chatID) {
			return nil
		}
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
	}, th.CommandEqual("changeTitle"))

}

func changeChatTitle(title string, chatID telego.ChatID, bh *th.BotHandler, ctx *th.Context) {
	err := ctx.Bot().SetChatTitle(ctx, &telego.SetChatTitleParams{ChatID: chatID, Title: title})
	if err != nil {
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "У меня нет прав на смену названия данного чата"})
	}
}

func ping(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		msgID := update.Message.MessageID
		var pingMsg string
		if len(update.Message.Text) >= 6 {
			pingMsg = update.Message.Text[6:]
		} else {
			pingMsg = update.Message.Text
		}
		chatID := update.Message.Chat.ChatID()
		if !fromChat(chatID) {
			return nil
		}
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: msgID, Quote: pingMsg}, ParseMode: telego.ModeMarkdownV2, ChatID: chatID, Text: "||" + strings.ReplaceAll(strings.Join(chat.Users, ", "), "_", "\\_") + "||"})

		return nil
	}, th.CommandEqual("ping"))
}

func addNewPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()

		if !fromChat(chatID) {
			return nil
		}
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		if len(update.Message.NewChatMembers) == 0 {
			return nil
		} else {
			newMembers := update.Message.NewChatMembers
			for _, newMember := range newMembers {
				if !newMember.IsBot {
					chat.Users = append(chat.Users, "@"+newMember.Username)
					write(chat, db)
					ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Новый юзер добавлен"})
				}
			}
		}
		return nil
	}, th.AnyMessage())
}

func delLeftPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()

		if !fromChat(chatID) {
			return nil
		}
		chat, err := read(int(chatID.ID), db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		}
		if update.Message.LeftChatMember == nil {
			return nil
		} else {
			if update.Message.LeftChatMember.IsBot {
				return nil
			}
			username := "@" + update.Message.LeftChatMember.Username
			leftUserIndex := slices.Index(chat.Users, username)
			chat.Users = slices.Delete(chat.Users, leftUserIndex, leftUserIndex+1)
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Старый юзер удален"})
			write(chat, db)
		}
		return nil
	}, th.AnyMessage())
}

func fromChat(id telego.ChatID) bool {
	if id.ID < 0 {
		return true
	}
	return false
}

func isAdmin(userId int, bot telego.Bot, ctx *th.Context, id telego.ChatID) bool {
	idList := []int{}
	admins, _ := bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{ChatID: id})
	for _, admin := range admins {
		idList = append(idList, int(admin.MemberUser().ID))
	}
	if slices.Contains(idList, userId) {
		return true
	}
	return false
}
