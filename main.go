// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var BotInstance *tgbotapi.BotAPI

func main() {
	//получение ключа
	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка чтения .env файла")
	}

	proxyURL, parseErr := url.Parse("http://127.0.0.1:2080")
	if parseErr != nil {
		log.Fatal("неверный URL прокси: ", parseErr)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// проврка БД
	if err = InitDB(); err != nil {
		log.Fatal("Ошибка подключения к БД: ", err)
	}
	defer DataBase.Close() // Гарантированно закроет базу

	//Инициализация ТГ бота.
	bot, err := tgbotapi.NewBotAPIWithClient(
		os.Getenv("ApiTelegramBot"),
		tgbotapi.APIEndpoint,
		httpClient,
	)
	if err != nil {
		log.Fatalf("Не удалось инициализировать бота: %v", err)
	}

	BotInstance = bot
	log.Printf("Бот успешно авторизован как @%s (ID: %d)", bot.Self.UserName, bot.Self.ID)

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
