// dataBase.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DataBase *sql.DB

func InitDB() error {
	//работа с информацией
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	//проверка все ли ок
	var err error
	DataBase, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("ошибка открытия соединения: %w", err)
	}

	// проверяем реальное соединение + повторные попытки
	for i := 0; i < 10; i++ {
		err = DataBase.Ping()
		if err == nil {
			return nil
		}
		log.Printf("⏳ Ожидание БД... попытка %d/10", i+1)
		time.Sleep(2 * time.Second)
	}

	//падаем с ошибкой
	return fmt.Errorf("не удалось подключиться к БД: %w", err)
}
