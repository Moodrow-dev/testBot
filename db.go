package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
)

func write(chat Chat, db *sql.DB) error {
	usersJson, err := json.Marshal(chat.Users)
	if err != nil {
		return err
	}
	err = deleteChat(chat.ID, db)
	err = createTable(db)
	if err != nil {
		return err
	}
	result, exec := db.Exec("INSERT INTO chats (id, main_topic, num, den, title, users) VALUES (?, ?, ?, ?, ?, ?)", chat.ID, chat.InfoThread, chat.Num, chat.Den, chat.Title, usersJson)
	if exec != nil {
		return exec
	}
	fmt.Println(result.LastInsertId())
	return nil
}

func read(id int64, db *sql.DB) (Chat, error) {
	// В запросе нужно добавить WHERE для фильтрации по id
	row := db.QueryRow("SELECT id, main_topic, num, den, title, users FROM chats WHERE id = ?", id)

	var (
		dbId      int64
		mainTopic int
		num       string
		den       string
		title     string
		usersJSON string
	)

	err := row.Scan(&dbId, &mainTopic, &num, &den, &title, &usersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return Chat{}, fmt.Errorf("Чат %d не найден", id)
		}
		return Chat{}, err
	}

	// Декодируем JSON массив пользователей
	var usersSlice []string
	if err := json.Unmarshal([]byte(usersJSON), &usersSlice); err != nil {
		return Chat{}, fmt.Errorf("failed to unmarshal users: %v", err)
	}

	// Создаем и возвращаем объект Chat
	chat := Chat{
		ID:         dbId,
		InfoThread: mainTopic,
		Num:        num,
		Den:        den,
		Title:      title,
		Users:      usersSlice,
	}

	return chat, nil
}

func pickOverIds(db *sql.DB) ([]int64, error) {
	raw, err := db.Query("SELECT id FROM chats")
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	var ids []int64

	for raw.Next() {
		var id int64

		err = raw.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func deleteChat(id int64, db *sql.DB) error {
	result, err := db.Exec("DELETE FROM chats WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat with id %d not found", id)
	}

	return nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY,
		main_topic INTEGER,
		num STRING,
		den STRING,
		title STRING,
		users STRING
    	name TEXT NOT NULL,
    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	return nil
}
