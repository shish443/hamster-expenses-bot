// handlers.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// время мск
var mskLoc = time.FixedZone("MSK", 10800)

// HandleHelp - приветствие и список команд
func HandleHelp() string {
	return "👋 Привет! Я бот для учёта расходов на хомячков.\n" +
		"📋 Команды:\n" +
		"• /add <сумма> [описание] - добавить расход (пример: /add 150 корм)\n" +
		"• /del <ID> - удалить запись по ID\n" +
		"• /calc [период] - посчитать расходы (день/неделя/месяц/квартал/полгода/год/все)\n" +
		"• /list - последние 5 записей\n" +
		"• /listWeek, /listMonth, /listQuarter и тд. -за период\n" +
		"• /info - об авторе"
}

// HandleAdd - добавление расхода
func HandleAdd(ctx context.Context, args []string, userID int64) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("укажите сумму: /add 150 [описание]")
	}

	// Парсинг суммы
	amount, err := strconv.ParseFloat(args[0], 64)
	if err != nil || amount <= 0 {
		return "", fmt.Errorf("сумма должна быть положительным числом")
	}

	// Описание
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	// запись в базу
	_, err = DataBase.ExecContext(
		ctx,
		"INSERT INTO expenses (user_id, amount, description) VALUES ($1, $2, $3)",
		userID, amount, desc,
	)
	if err != nil {
		log.Printf("❌ DB Error (Add): %v | user=%d, amount=%.2f", err, userID, amount)
		return "", fmt.Errorf("не удалось сохранить расход. Попробуйте позже.")
	}
	return fmt.Sprintf("✅ Добавлено: %.2f ₽ [%s]", amount, desc), nil
}

// HandleDelete - удаление записи
func HandleDelete(ctx context.Context, args []string, userID int64) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("укажите ID записи: /del 123")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("ID должен быть числом")
	}

	// удаление
	res, err := DataBase.Exec(
		"DELETE FROM expenses WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		log.Printf("DB Error (Del): %v", err)
		return "", fmt.Errorf("ошибка удаления записи")
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return "", fmt.Errorf("запись #%d не найдена или не принадлежит вам", id)
	}

	return fmt.Sprintf("🗑️ Запись #%d удалена", id), nil
}

// HandleCalc -подсчёт расходов за период
func HandleCalc(ctx context.Context, args []string, userID int64) (string, error) {
	period := "all"
	if len(args) > 0 {
		period = strings.ToLower(args[0])
	}

	// Рассчитываем диапазон дат
	now := time.Now().In(mskLoc)
	var startDate time.Time

	switch period {
	case "день", "day":
		startDate = now.AddDate(0, 0, -1)
	case "неделя", "week":
		startDate = now.AddDate(0, 0, -7)
	case "месяц", "month":
		startDate = now.AddDate(0, -1, 0)
	case "квартал", "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "полгода", "halfyear", "6months":
		startDate = now.AddDate(0, -6, 0)
	case "год", "year":
		startDate = now.AddDate(-1, 0, 0)
	case "все", "all", "":
		return calcAllTime(userID)
	default:
		return "", fmt.Errorf("неизвестный период: %s. Доступные: день/неделя/месяц/квартал/полгода/год/все", period)
	}

	// Запрос суммы за период
	var total float64
	err := DataBase.QueryRowContext(
		ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1 AND created_at >= $2",
		userID, startDate,
	).Scan(&total)

	if err != nil {
		log.Printf("❌ DB Error (Calc): %v | user=%d, period=%s", err, userID, period)
		return "", fmt.Errorf("ошибка подсчёта расходов")
	}

	return fmt.Sprintf("📊 Расходы за %s: %.2f ₽", period, total), nil
}

// Вспомогательная функция
func calcAllTime(userID int64) (string, error) {
	var total sql.NullFloat64
	err := DataBase.QueryRow(
		"SELECT SUM(amount) FROM expenses WHERE user_id = $1",
		userID,
	).Scan(&total)

	if err != nil {
		return "", fmt.Errorf("ошибка подсчёта: %w", err)
	}

	if !total.Valid || total.Float64 == 0 {
		return "📊 Расходы за всё время: 0 ₽ (записей нет)", nil
	}

	return fmt.Sprintf("📊 Расходы за всё время: %.2f ₽", total.Float64), nil
}

// HandleList - вывод записей
func HandleList(ctx context.Context, args []string, userID int64, listType string) (string, error) {
	// безопасный запрос.
	rows, err := DataBase.QueryContext(
		ctx,
		`SELECT id, amount, description, created_at FROM expenses 
		 WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100`,
		userID,
	)

	if err != nil {
		log.Printf("❌ DB Error (List): %v | user=%d", err, userID)
		return "", fmt.Errorf("ошибка чтения записей")
	}
	defer rows.Close()

	// Рассчитываем дату начала периода
	now := time.Now().In(mskLoc)
	var startDate time.Time
	switch listType {
	case "list":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "halfyear":
		startDate = now.AddDate(0, -6, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	case "all":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = now.AddDate(-100, 0, 0)
	}

	var result strings.Builder
	result.WriteString("📋 Ваши записи:\n")
	count := 0

	for rows.Next() {
		var id int
		var amount float64
		var description sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &amount, &description, &createdAt); err != nil {
			continue
		}

		if createdAt.Before(startDate) {
			continue
		}

		desc := "без описания"
		if description.Valid && description.String != "" {
			desc = description.String
		}

		result.WriteString(fmt.Sprintf("%d. %.2f ₽ — %s (%s)\n",
			id, amount, desc, createdAt.In(mskLoc).Format("02.01 15:04")))
		count++

		if listType == "list" && count >= 5 {
			break
		}
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("ошибка при чтении данных")
	}

	if count == 0 {
		return "📭 Записей за этот период нет. Добавьте первую через /add 150 корм", nil
	}
	return result.String(), nil
}

// инфо об авторе
func InfoAuthor() string {
	return "👨‍💻 Бот сделан: https://github.com/shish443\n" +
		"🔧 Стек: Go + PostgreSQL + Docker\n" +
		"💡 Это мой первый пет-проект. Критика приветствуется!"
}
