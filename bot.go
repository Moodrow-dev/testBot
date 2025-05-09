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

func createBotAndPoll() (*telego.Bot, th.BotHandler, error) {
	err := godotenv.Load("token.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
		return nil, th.BotHandler{}, err
	}
	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	return bot, *bh, nil
}

func chatInit(bh *th.BotHandler, db *sql.DB) Chat {
	var chat Chat
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		if !fromChat(chatID) {
			return nil
		}
		var err error
		chat, err = read(chatID.ID, db)
		if err != nil {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: err.Error()})
			chat = createChat(chatID.ID, update.Message.Chat.Title)
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
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat, err := read(chatID.ID, db)
		if err != nil {
			return nil
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
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat, err := read(chatID.ID, db)
		if err != nil {
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
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf("Список пользователей\n%v", chat.Users)})
			} else {
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Список пользователей очищен"})
			}
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
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat, err := read(chatID.ID, db)
		if err != nil {
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
	}, th.CommandEqual("changeWeek"))
}

func changeTitle(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat, err := getChatByID(chatID, db, ctx.Bot())
		if err != nil {
			return nil
		}
		title := strings.Split(update.Message.Text, " ")[1]
		chat.Title = title
		title = fmt.Sprintf("[%v] %v", chat.Num, title)
		changeChatTitle(title, chatID, ctx.Bot())
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Название изменено успешно"})
		write(chat, db)
		return nil
	}, th.CommandEqual("changeTitle"))

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
		chat, err := getChatByID(chatID, db, ctx.Bot())
		if err != nil {
			return nil
		}
		users := chat.Users
		if len(users) <= 1 {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Ошибка: некого пинговать"})
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: msgID, Quote: pingMsg}, ParseMode: telego.ModeMarkdownV2, ChatID: chatID, Text: "||" + strings.ReplaceAll(strings.Join(users, ", "), "_", "\\_") + "||"})
		}
		return nil
	}, th.CommandEqual("ping"))
}

// Пассивные функции \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\
func addNewPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat, err := getChatByID(chatID, db, ctx.Bot())
		if err != nil {
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
		chat, err := getChatByID(chatID, db, ctx.Bot())
		if err != nil {
			return nil
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

func changeWeekMain(chatID telego.ChatID, bot *telego.Bot, db *sql.DB) error {
	ctx := context.Background()
	chat, err := read(chatID.ID, db)
	if err != nil {
		return err
	}

	// Получаем текущее название чата
	chatInfo, err := bot.GetChat(ctx, &telego.GetChatParams{ChatID: chatID})
	if err != nil {
		return err
	}
	oldTitle := chatInfo.Title

	numTitle := fmt.Sprintf("[%v] %v", chat.Den, chat.Title)
	denTitle := fmt.Sprintf("[%v] %v", chat.Num, chat.Title)

	// Определяем какое название установить
	newTitle := denTitle
	if oldTitle == numTitle {
		newTitle = denTitle
	} else {
		newTitle = numTitle
	}

	// Меняем название
	err = bot.SetChatTitle(ctx, &telego.SetChatTitleParams{
		ChatID: chatID,
		Title:  newTitle,
	})
	return err
}
