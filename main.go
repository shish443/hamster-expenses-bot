package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка чтения .env файла")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("ApiTelegramBot"))
	if err != nil {
		log.Panic(err)
	}

	upTime := tgbotapi.NewUpdate(0)
	upTime.Timeout = 60
	updates := bot.GetUpdatesChan(upTime)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.Chat.ID, update.Message.Text)
			messageTg := tgbotapi.NewMessage(update.Message.From.ID, update.Message.Text)
			messageTg.ReplyToMessageID = update.Message.MessageID

			bot.Send(messageTg)
		}

	}

}
