package main

import "fmt"

type Chat struct {
	ID        int64    `bson:"id"`
	MainTopic int      `bson:"main_topic"`
	Num       string   `bson:"num"`
	Den       string   `bson:"den"`
	Title     string   `bson:"title"`
	Users     []string `bson:"users"`
}

func createChat(id int64, title string) Chat {
	chat := Chat{}
	chat.ID = id
	chat.MainTopic = 1
	chat.Num = "Числитель"
	chat.Den = "Знаменатель"
	chat.Title = title
	chat.Users = []string{}
	return chat
}

func (c Chat) toString() string {
	return fmt.Sprintf("%v %v %v %v %v %v", c.ID, c.MainTopic, c.Num, c.Den, c.Title, c.Users)
}
