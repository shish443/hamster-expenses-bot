// botLoop.go
package main

import (
	"context"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	nonCommandCounter = struct {
		sync.Mutex
		m map[int64]int //счётчик подряд идущих не команд
	}{
		m: make(map[int64]int),
	}
)

func StartBotLoop(ctx context.Context, bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update) {
	for update := range updates {
		//Обработку сообщения в отдельной горутине
		go func(u tgbotapi.Update) {
			incrementMessages()
			//проверка на пустое сообщение и бота
			if u.Message == nil || u.Message.Text == "" || u.Message.From.IsBot == true {
				return
			}

			//проверка длинны
			if len(u.Message.Text) > 4096 {
				sendReply(ctx, bot, u.Message.Chat.ID, u.Message.MessageID, "✂️ Сообщение слишком длинное. Пожалуйста, напиши короче.")
				return
			}

			//логика не команд
			isCommand := strings.HasPrefix(strings.TrimSpace(u.Message.Text), "/")

			if !isCommand {
				shouldReply := handleNonCommand(u.Message.From.ID)
				if !shouldReply {
					return
				}
				//Если должны ответить-покажем подсказку
			} else {
				resetNonCommandCounter(u.Message.From.ID)
			}
			cmd, err := Parser(bot, u)

			// parser не распознал команду
			if err != nil {
				if err.Error() == "не команда" {
					sendReply(ctx, bot, u.Message.Chat.ID, u.Message.MessageID,
						"чтобы добавить расход используй команду:\n\n"+
							"`/add 150 корм`\n\n"+
							"Напиши /help, чтобы посмотреть все команды.",
					)
					return
				}

				// Другие ошибки парсера
				log.Printf("Parser error: %v (text: %q)", err, u.Message.Text)
				sendReply(ctx, bot, u.Message.Chat.ID, u.Message.MessageID,
					"не понял команду. Введите /help")
				return
			}

			if cmd == nil {
				return
			}

			// все ок, передаем команду
			response, err := Router(ctx, bot, cmd)
			if err != nil {
				incrementErrors()
				log.Printf("ошибка роутера: %v (cmd: %+v)", err, cmd)
				sendReply(ctx, bot, cmd.ChatID, cmd.MsgID, "Произошла внутренняя ошибка. Попробуйте позже.")
				return
			}

			//все ок отправляем сообщение
			if err := sendReply(ctx, bot, cmd.ChatID, cmd.MsgID, response); err != nil {
				log.Printf("Failed to send reply: %v", err)
			}

			log.Printf("Handled: %s from UserID %d", cmd.Name, cmd.UserID)

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

	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyToMessageID = replyTo

	_, err := bot.Send(msg)
	return err
}

func handleNonCommand(userID int64) bool {
	nonCommandCounter.Lock()
	defer nonCommandCounter.Unlock()

	nonCommandCounter.m[userID]++
	count := nonCommandCounter.m[userID]

	// Отвечаем 1 раз из 5
	return count%5 == 0
}

func resetNonCommandCounter(userID int64) {
	nonCommandCounter.Lock()
	delete(nonCommandCounter.m, userID) // сбрасываем при любой команде
	nonCommandCounter.Unlock()
}
