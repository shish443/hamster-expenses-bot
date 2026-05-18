// botLoop.go
package main

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBotLoop(ctx context.Context, bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update) {
	for update := range updates {
		//Обработку сообщения в отдельной горутине
		go func(u tgbotapi.Update) {
			//проверка на пустое сообщение и бота
			if u.Message == nil || u.Message.Text == "" || u.Message.From.IsBot == true {
				return
			}

			//проверка длинны
			if len(u.Message.Text) > 4096 {
				sendReply(ctx, bot, u.Message.Chat.ID, u.Message.MessageID, "✂️ Сообщение слишком длинное. Пожалуйста, напиши короче.")
				return
			}

			cmd, err := Parser(bot, u)
			// parser не распознал команду
			if err != nil {
				log.Printf("⚠️ Parser error: %v (text: %q)", err, u.Message.Text)
				sendReply(ctx, bot, u.Message.Chat.ID, u.Message.MessageID, "❌ Не понял команду. Введите /help")
				return
			}

			if cmd == nil {
				return
			}
			// все ок, передаем команду
			response, err := Router(ctx, cmd)
			if err != nil {
				log.Printf("❌ Router error: %v (cmd: %+v)", err, cmd)
				sendReply(ctx, bot, cmd.ChatID, cmd.MsgID, "⚠️ Произошла внутренняя ошибка. Попробуйте позже.")
				return
			}

			//все ок отправляем сообщение
			if err := sendReply(ctx, bot, cmd.ChatID, cmd.MsgID, response); err != nil {
				log.Printf("❌ Failed to send reply: %v", err)
			}

			log.Printf("✅ Handled: %s from UserID %d", cmd.Name, cmd.UserID)

			// Сигнал завершения=выходим из цикла
			select {
			case <-ctx.Done():
				return
			default:
			}
		}(update)
	}
}

func sendReply(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, replyTo int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyTo
	_, err := bot.Send(msg)
	return err
}
