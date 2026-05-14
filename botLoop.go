// botLoop.go
package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBotLoop(bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update) {
	for update := range updates {
		//проверка на пустое сообщение и бота
		if update.Message == nil || update.Message.Text == "" || update.Message.From.IsBot == true {
			continue
		}
		//проверка длинны
		if len(update.Message.Text) > 4096 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✂️ Сообщение слишком длинное. Пожалуйста, напиши короче.")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)

			continue
		}
		cmd, err := Parser(bot, update)
		// parser не распознал команду
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Не понял команду. Введите /help")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)

			continue
		}

		// все ок, передаем команду
		response, err := Router(cmd)
		if err != nil {
			msg := tgbotapi.NewMessage(cmd.ChatID, fmt.Sprintf("⚠️ Ошибка: %v", err))
			msg.ReplyToMessageID = cmd.MsgID
			bot.Send(msg)

			continue
		}

		//все ок отправляем сообщение
		msg := tgbotapi.NewMessage(cmd.ChatID, response)
		msg.ReplyToMessageID = cmd.MsgID
		bot.Send(msg)

		log.Printf("📩 От %d: %s", update.Message.From.ID, update.Message.Text)
		log.Println(cmd)

	}
}
