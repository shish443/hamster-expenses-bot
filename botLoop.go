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
			//обработка инлайн кнопок
			if u.CallbackQuery != nil {
				handleCallbackQuery(ctx, bot, u.CallbackQuery)
				return
			}
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
			if err := sendReplyWithKeyboard(ctx, bot, cmd.ChatID, cmd.MsgID, response); err != nil {
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

	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = replyTo

	_, err := bot.Send(msg)
	return err
}

// отправляет ответ + главное меню с кнопками
func sendReplyWithKeyboard(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, replyTo int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = replyTo

	msg.ReplyMarkup = getMainKeyboard()

	_, err := bot.Send(msg)
	return err
}

func sendReplyWithCalcKeyboard(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, replyTo int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyToMessageID = replyTo
	msg.ReplyMarkup = getCalcKeyboard()

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

func handleCallbackQuery(ctx context.Context, bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	_, _ = bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	switch cb.Data {
	case "cmd_help":
		sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, HandleHelp())
		return

	case "cmd_add":
		sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID,
			"Напиши расход, например:\n\n`/add 150 корм`")
		return

	case "cmd_calc":
		sendReplyWithCalcKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, "Выбери период:")
		return

	case "cmd_list", "cmd_del":
		// Показываем список с кнопками удаления
		text, keyboard, err := buildListWithKeyboard(ctx, cb.From.ID, "list")
		if err != nil {
			sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, "❌ Ошибка при получении списка")
			return
		}
		if text == "📭 У тебя пока нет записей расходов.\n\nДобавь первую с помощью кнопки «➕ Добавить» или команды /add" {
			sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, text)
			return
		}

		msg := tgbotapi.NewMessage(cb.Message.Chat.ID, text)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
		return

	case "calc_day", "calc_week", "calc_month", "calc_quarter", "calc_halfyear", "calc_year", "calc_all":
		period := cb.Data[5:]
		response, err := HandleCalc(ctx, []string{period}, cb.From.ID)
		if err != nil {
			response = "❌ Ошибка: " + err.Error()
		}
		sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, response)
		return

	default:
		// Удаление записи
		if len(cb.Data) > 4 && cb.Data[:4] == "del_" {
			idStr := cb.Data[4:]
			response, err := HandleDelete(ctx, []string{idStr}, cb.From.ID)
			if err != nil {
				response = "❌ " + err.Error()
			}
			sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, response)
			return
		}

		// Главное меню
		if cb.Data == "main_menu" {
			sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, "Главное меню:")
			return
		}

		sendReplyWithKeyboard(ctx, bot, cb.Message.Chat.ID, cb.Message.MessageID, "Неизвестная кнопка")
	}
}
