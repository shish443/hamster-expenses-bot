// handlers.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// время мск
var mskLoc = time.FixedZone("MSK", 3*60*60) // ютк+3

// HandleHelp - приветствие и список команд
// HandleHelp - приветствие и список команд
func HandleHelp() string {
	return `👋 Привет! Я бот для учёта расходов на хомячков 🐹

📋 Доступные команды:

• /add <сумма> [описание] — добавить расход
  Пример: /add 150 корм

• /del <ID> — удалить запись
  Пример: /del 42

• /calc [период] — сумма расходов
  Примеры: /calc day, /calc week, /calc month, /calc all

• /list — последние 5 записей
• /listWeek, /listMonth, /listQuarter, /listYear — записи за период

• /complaint <текст> — отправить жалобу админу

• /info — об авторе
• /help — это сообщение

Просто нажми на команду выше — она скопируется! 👆`
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

	// Ограничение на максимальную сумму (защита от случайных/злонамеренных больших чисел)
	if amount > 1_000_000 {
		return "", fmt.Errorf("слишком большая сумма. Максимум 1 000 000 ₽ за одну запись")
	}

	// Описание
	desc := ""
	if len(args) > 1 {
		desc = strings.Join(args[1:], " ")
	}

	if len(desc) > 200 {
		return "", fmt.Errorf("описание слишком длинное (макс. 200 символов)")
	}

	// запись в базу
	tx, err := DataBase.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("BeginTx failed: %v", err)
		return "", fmt.Errorf("внутренняя ошибка сервера")
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO expenses (user_id, amount, description, created_at) 
		 VALUES ($1, $2, $3, NOW() AT TIME ZONE 'UTC')`,
		userID, amount, desc,
	)
	if err != nil {
		log.Printf("ошибка добавления в базу данных: %v | user=%d, amount=%.2f", err, userID, amount)
		return "", fmt.Errorf("не удалось сохранить расход. Попробуйте позже.")
	}

	if err = tx.Commit(); err != nil {
		log.Printf("ошибка: %v", err)
		return "", fmt.Errorf("не удалось сохранить расход")
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
	case "день", "day", "д", "d":
		startDate = now.AddDate(0, 0, -1)
	case "week", "неделя", "н", "w":
		startDate = now.AddDate(0, 0, -7)
	case "month", "месяц", "м", "m":
		startDate = now.AddDate(0, -1, 0)
	case "quarter", "квартал", "к", "q":
		startDate = now.AddDate(0, -3, 0)
	case "halfyear", "полгода", "п", "h", "6months", "6 months", "6месяцев", "6 месяцев":
		startDate = now.AddDate(0, -6, 0)
	case "year", "год", "г", "y":
		startDate = now.AddDate(-1, 0, 0)
	case "all", "алл", "все время", "a", "все":
		return calcAllTime(ctx, userID)
	default:
		return "", fmt.Errorf("неизвестный период: %s. Доступные: день, неделя, месяц, квартал, полгода, год, все", period)
	}

	// Запрос суммы за период кроме всего времени
	var total float64
	err := DataBase.QueryRowContext(
		ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1 AND created_at >= $2",
		userID, startDate,
	).Scan(&total)

	if err != nil {
		log.Printf("DB Error (Calc): %v | user=%d, period=%s", err, userID, period)
		return "", fmt.Errorf("ошибка подсчёта расходов")
	}

	return fmt.Sprintf("📊 Расходы за %s: %.2f ₽", period, total), nil
}

// Вспомогательная функция
func calcAllTime(ctx context.Context, userID int64) (string, error) {
	var total float64

	err := DataBase.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1",
		userID,
	).Scan(&total)

	if err != nil {
		log.Printf("❌ DB Error (calcAllTime): %v | user=%d", err, userID)
		return "", fmt.Errorf("ошибка подсчёта расходов")
	}

	if total == 0 {
		return "📊 Расходы за всё время: 0 ₽ (записей пока нет)", nil
	}

	return fmt.Sprintf("📊 Расходы за всё время: %.2f ₽", total), nil
}

// HandleList - вывод записей
func HandleList(ctx context.Context, args []string, userID int64, listType string) (string, error) {
	text, _, err := buildListWithKeyboard(ctx, userID, listType)
	return text, err
}

// возвращает текст + клавиатуру
func buildListWithKeyboard(ctx context.Context, userID int64, listType string) (string, tgbotapi.InlineKeyboardMarkup, error) {
	limit := 10
	now := time.Now().In(mskLoc)
	var startDate time.Time

	switch listType {
	case "list":
		limit = 5
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	case "week", "неделя", "н", "w":
		startDate = now.AddDate(0, 0, -7)
	case "month", "месяц", "м", "m":
		startDate = now.AddDate(0, -1, 0)
	case "quarter", "квартал", "к", "q":
		startDate = now.AddDate(0, -3, 0)
	case "halfyear", "полгода", "п", "h":
		startDate = now.AddDate(0, -6, 0)
	case "year", "год", "г", "y":
		startDate = now.AddDate(-1, 0, 0)
	case "all", "алл", "все время", "a", "все":
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		startDate = now.AddDate(-100, 0, 0)
	}

	rows, err := DataBase.QueryContext(
		ctx,
		`SELECT id, amount, description, created_at 
		 FROM expenses 
		 WHERE user_id = $1 
		   AND created_at >= $2 
		 ORDER BY created_at DESC 
		 LIMIT $3`,
		userID, startDate, limit,
	)
	if err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, fmt.Errorf("ошибка чтения записей")
	}
	defer rows.Close()

	var result strings.Builder
	result.WriteString("📋 Ваши последние записи:\n\n")
	count := 0

	var records []struct {
		ID   int
		Text string
	}

	for rows.Next() {
		var id int
		var amount float64
		var description sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &amount, &description, &createdAt); err != nil {
			continue
		}

		desc := "без описания"
		if description.Valid && description.String != "" {
			desc = description.String
		}

		line := fmt.Sprintf("%d. %.2f ₽ — %s (%s)\n",
			id, amount, desc, createdAt.In(mskLoc).Format("02.01 15:04"))

		result.WriteString(line)
		records = append(records, struct {
			ID   int
			Text string
		}{id, line})
		count++
	}

	if err = rows.Err(); err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, err
	}

	if count == 0 {
		return "📭 У тебя пока нет записей расходов.\n\nДобавь первую с помощью кнопки «➕ Добавить» или команды /add", tgbotapi.InlineKeyboardMarkup{}, nil
	}

	keyboard := createDeleteKeyboard(records)

	return result.String() + "\n🗑 Нажми на кнопку, чтобы удалить запись:", keyboard, nil
}

// инфо об авторе
func InfoAuthor() string {
	return "👨‍💻 Бот сделан: https://github.com/shish443\n" +
		"🔧 Стек: Go + PostgreSQL + Docker\n" +
		"💡 Это мой первый пет-проект. Критика приветствуется!"
}

func HandleComplaint(ctx context.Context, bot *tgbotapi.BotAPI, args []string, userID int64) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("напишите текст жалобы. Пример: /complaint calc считает неправильно")
	}

	text := strings.Join(args, " ")

	_, err := DataBase.ExecContext(ctx,
		"INSERT INTO complaints (user_id, description) VALUES ($1, $2)",
		userID, text,
	)
	if err != nil {
		log.Printf("DB Error (Complaint): %v", err)
		incrementErrors()
		return "", fmt.Errorf("не удалось сохранить жалобу. Попробуйте позже.")
	}

	incrementComplaints()
	//уведомление админу
	notifyAdmin(ctx, bot, fmt.Sprintf("новая жалоба от пользователя %d:\n\n%s", userID, text))

	return "✅ Жалоба отправлена админу. Спасибо!", nil
}

func notifyAdmin(ctx context.Context, bot *tgbotapi.BotAPI, text string) {
	adminID := getAdminID()
	if adminID == 0 || bot == nil {
		return
	}
	if err := sendReply(ctx, bot, adminID, 0, text); err != nil {
		log.Printf("не удалось отправить уведомление админу: %v", err)
	}
}

func HandleAdmin(ctx context.Context, args []string, userID int64) (string, error) {
	adminID := getAdminID()
	if userID != adminID {
		return "Доступ запрещён. Вы не администратор.", nil
	}

	if len(args) == 0 {
		return getGlobalStats(), nil
	}

	// Статистика по пользователю
	targetID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return "", fmt.Errorf("использование: /admin или /admin <user_id>")
	}

	return getUserStats(ctx, targetID)
}

func getAdminID() int64 {
	id, _ := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	return id
}

func HandleAdminHelp() string {
	return `Админ-панель

/admin - общая статистика бота
/admin <user_id> - статистика конкретного пользователя
/adminhelp - это сообщение`
}

func getGlobalStats() string {
	var totalUsers int
	DataBase.QueryRow("SELECT COUNT(DISTINCT user_id) FROM expenses").Scan(&totalUsers)

	var totalExpenses float64
	DataBase.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses").Scan(&totalExpenses)

	return fmt.Sprintf(`📊 Глобальная статистика бота

⏱ Uptime: %s
📨 Всего сообщений: %d
❌ Ошибок: %d
📢 Жалоб: %d
👥 Уникальных пользователей: %d
💰 Общая сумма расходов: %.2f ₽`,
		getUptime(), totalMessages, totalErrors, totalComplaints, totalUsers, totalExpenses)
}

func getUserStats(ctx context.Context, userID int64) (string, error) {
	var count int
	var sum float64
	DataBase.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM expenses WHERE user_id = $1",
		userID).Scan(&count, &sum)

	var complaints int
	DataBase.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM complaints WHERE user_id = $1",
		userID).Scan(&complaints)

	return fmt.Sprintf(`Статистика пользователя %d

📝 Записей расходов: %d
💰 Сумма: %.2f ₽
📢 Жалоб отправлено: %d`, userID, count, sum, complaints), nil
}
