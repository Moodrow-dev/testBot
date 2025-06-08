package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"os"
	"slices"
	"strings"
)

func CreateBotAndPoll() (*telego.Bot, *th.BotHandler, error) {
	err := godotenv.Load("token.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	return bot, bh, nil
}

func ChatInit(bh *th.BotHandler, db *sql.DB) *Chat {
	var chat *Chat
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		if !fromChat(chatID) {
			return nil
		}
		chat = read(chatID.ID, db)
		if chat == nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Бот не инициализирован"})
			chat = CreateChat(chatID.ID, update.Message.Chat.Title)
			write(chat, db)
		}

		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Чат успешно инициализирован"})
		return nil
	}, th.CommandEqual("init"))
	return chat
}

func ChangeNumDenum(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}

		args := strings.Split(update.Message.Text, " ")
		num := args[1]
		denum := args[2]
		chat.Num = num
		chat.Den = denum
		write(chat, db)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: fmt.Sprintf("Успешно.\nЧислитель теперь: %v,\nзнаменатель теперь: %v", num, denum)})
		return nil
	}, th.CommandEqualArgc("changeWeekTitle", 2))
}

func SetUsers(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}

		people := strings.Split(strings.ReplaceAll(update.Message.Text, ",", ""), " ")[1:]
		flag := true
		for _, peopleStr := range people {
			if peopleStr[0] != '@' {
				flag = false
			}
		}
		if flag {
			chat.Users = []string{}
			chat.Users = append(chat.Users, people...)
			if len(chat.Users) != 0 {
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: fmt.Sprintf("Список пользователей\n%v", chat.Users)})
			} else {
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Список пользователей очищен"})
			}
			write(chat, db)
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Неверный формат команды!"})
		}
		return nil
	}, th.CommandEqual("SetUsers"))
}

func ChangeWeek(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}
		oldTitle := update.Message.Chat.Title
		numTitle := fmt.Sprintf("[%v] %v", chat.Den, chat.Title)
		denTitle := fmt.Sprintf("[%v] %v", chat.Num, chat.Title)
		if oldTitle != numTitle {
			changeChatTitle(numTitle, chatID, ctx.Bot())
		} else {
			changeChatTitle(denTitle, chatID, ctx.Bot())
		}
		return nil
	}, th.CommandEqual("ChangeWeek"))
}

func ChangeTitle(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, db, ctx.Bot())
		if chat == nil {
			return nil
		}

		title := strings.Join(strings.Split(update.Message.Text, " ")[1:], " ")
		chat.Title = title
		title = fmt.Sprintf("[%v] %v", chat.Num, title)
		changeChatTitle(title, chatID, ctx.Bot())
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, DisableNotification: true, Text: "Название изменено успешно"})
		write(chat, db)
		return nil
	}, th.CommandEqual("ChangeTitle"))

}

func Ping(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		msgID := update.Message.MessageID
		var pingMsg string
		if len(update.Message.Text) >= 6 {
			pingMsg = update.Message.Text[6:]
		} else {
			pingMsg = update.Message.Text
		}
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, ctx.Bot())
		if chat == nil {
			return nil
		}

		users := chat.Users
		if len(users) <= 1 {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Ошибка: некого пинговать"})
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: msgID, Quote: pingMsg}, ParseMode: telego.ModeMarkdownV2, ChatID: chatID, Text: "||" + strings.ReplaceAll(strings.Join(users, ", "), "_", "\\_") + "||"})
		}
		return nil
	}, th.CommandEqual("Ping"))
}

// Пассивные функции \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\
func AddNewPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, ctx.Bot())
		if chat == nil {
			return nil
		}

		if len(update.Message.NewChatMembers) == 0 {
			return nil
		} else {
			newMembers := update.Message.NewChatMembers
			for _, newMember := range newMembers {
				if !newMember.IsBot {
					chat.Users = append(chat.Users, "@"+newMember.Username)
					write(chat, db)
					ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Новый юзер добавлен"})
				}
			}
		}
		return nil
	}, th.AnyMessage())
}

func DelLeftPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, ctx.Bot())
		bot := ctx.Bot()
		if chat == nil {
			return nil
		}
		chat = getChatByID(chatID, db, bot)
		if update.Message.LeftChatMember == nil {
			return nil
		} else {
			if update.Message.LeftChatMember.IsBot {
				return nil
			}
			username := "@" + update.Message.LeftChatMember.Username
			leftUserIndex := slices.Index(chat.Users, username)
			chat.Users = slices.Delete(chat.Users, leftUserIndex, leftUserIndex+1)
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Старый юзер удален"})
			write(chat, db)
		}
		return nil
	}, th.AnyMessage())
}

func AdvertiseGit(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, db, bot)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на звездочет](https://github.com/voskhod-1/starsresearch)"})
		return nil
	}, th.And(th.Or(th.TextContains("гит"), th.TextContains("звездочет"), th.TextContains("космо")), th.Not(th.Or(th.TextContains("гитлер"), th.TextContains("гитар")))))
}

func AdvertiseBoosty(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, db, bot)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на бусти](https://boosty.to/starsresearch)"})
		return nil
	}, th.TextContains("бусти"))
}

func SetMainThread(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, ctx.Bot())
		if chat == nil {
			return nil
		}
		threadID := update.Message.MessageThreadID
		chat.InfoThread = threadID
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Тема для уведомлений установлена"})
		write(chat, db)
		return nil
	}, th.CommandEqual("SetMainThread"))
}
