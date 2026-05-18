// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	//получение ключа
	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка чтения .env файла")
	}

	// проврка БД
	if err := InitDB(); err != nil {
		log.Fatal("❌ Ошибка подключения к БД: ", err)
	}
	defer DataBase.Close() // Гарантированно закроет базу

	//Инициализация ТГ бота.
	bot, err := tgbotapi.NewBotAPI(os.Getenv("ApiTelegramBot"))
	if err != nil {
		log.Panic(err)
	}

	//проверяет новые сообщение
	upTime := tgbotapi.NewUpdate(0)
	upTime.Timeout = 60
	updates := bot.GetUpdatesChan(upTime)

	//завершаем работу
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go StartBotLoop(ctx, bot, updates)

	log.Println("Бот работает.")
	<-ctx.Done()
	log.Println("Корректное выключение")
}
