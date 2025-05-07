package main

import (
	"database/sql"
	"errors"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"slices"
)

func changeChatTitle(title string, chatID telego.ChatID, bh *th.BotHandler, ctx *th.Context) {
	err := ctx.Bot().SetChatTitle(ctx, &telego.SetChatTitleParams{ChatID: chatID, Title: title})
	if err != nil {
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "У меня нет прав на смену названия данного чата"})
	}
}

func fromChat(id telego.ChatID) bool {
	if id.ID < 0 {
		return true
	}
	return false
}

func isAdmin(userId int, ctx *th.Context, id telego.ChatID) bool {
	idList := []int{}
	bot := ctx.Bot()
	admins, _ := bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{ChatID: id})
	for _, admin := range admins {
		idList = append(idList, int(admin.MemberUser().ID))
	}
	if slices.Contains(idList, userId) {
		return true
	}
	bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: id, Text: "Ошибка! У вас недостаточно прав для выполнения этой команды"})
	return false
}

func getChatByID(chatID telego.ChatID, db *sql.DB, ctx *th.Context) (Chat, error) {
	if !fromChat(chatID) {
		return Chat{}, errors.New("Не из чата!")
	}
	chat, err := read(int(chatID.ID), db)
	if err != nil {
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
	}
	return chat, nil
}
