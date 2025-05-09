package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mymmrac/telego"
	"slices"
)

func changeChatTitle(title string, chatID telego.ChatID, bot *telego.Bot) {
	ctx := context.Background()
	err := bot.SetChatTitle(ctx, &telego.SetChatTitleParams{ChatID: chatID, Title: title})
	if err != nil {
		bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "У меня нет прав на смену названия данного чата"})
	}
}

func fromChat(id telego.ChatID) bool {
	if id.ID < 0 {
		return true
	}
	return false
}

func isAdmin(userId int64, bot *telego.Bot, id telego.ChatID) bool {
	ctx := context.Background()
	idList := []int64{}
	admins, _ := bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{ChatID: id})
	for _, admin := range admins {
		idList = append(idList, admin.MemberUser().ID)
	}
	if slices.Contains(idList, userId) {
		return true
	}
	bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: id, Text: "Ошибка! У вас недостаточно прав для выполнения этой команды"})
	return false
}

func getChatByID(chatID telego.ChatID, db *sql.DB, bot *telego.Bot) (Chat, error) {
	ctx := context.Background()
	if !fromChat(chatID) {
		return Chat{}, errors.New("Не из чата!")
	}
	chat, err := read(chatID.ID, db)
	if err != nil {
		bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
	}
	return chat, nil
}
