package main

import "fmt"

type Chat struct {
	ID         int64    `bson:"id"`
	InfoThread int      `bson:"main_topic"`
	Num        string   `bson:"num"`
	Den        string   `bson:"den"`
	Title      string   `bson:"title"`
	Users      []string `bson:"users"`
}

func CreateChat(id int64, title string) *Chat {
	chat := Chat{}
	chat.ID = id
	chat.InfoThread = 0
	chat.Num = "Числитель"
	chat.Den = "Знаменатель"
	chat.Title = title
	chat.Users = []string{}
	return &chat
}

func (c Chat) ToString() string {
	return fmt.Sprintf("%v %v %v %v %v %v", c.ID, c.InfoThread, c.Num, c.Den, c.Title, c.Users)
}
