// parser.go
package main

import (
	"errors"
	"strings"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Parser(bot *tgbotapi.BotAPI, update tgbotapi.Update) (*Command, error) {
	// Если это не команда бот игнорит
	if update.Message == nil || update.Message.Text == "" {
		return nil, nil
	}
	// удаление пробелов
	text := strings.TrimSpace(update.Message.Text)

	// удаление / и определение команда ли это. работа в юникоде, да излишество, но учимся сразу правильно
	firstRune, size := utf8.DecodeRuneInString(text)

	if firstRune == '/' {
		text = text[size:]
		text = strings.TrimSpace(text)
		if text == "" {
			return nil, nil
		}

	} else {
		return nil, errors.New("не команда")
	}

	//разбиваем строку
	parts := strings.Fields(text)
	if len(parts) == 0 { // на всякий
		return nil, errors.New("пустая команда")
	}

	return &Command{
		Name:   strings.ToLower(parts[0]), // "add"
		Args:   parts[1:],                 // ["150", "корм"]
		UserID: update.Message.From.ID,
		ChatID: update.Message.Chat.ID,
		MsgID:  update.Message.MessageID,
	}, nil
}
